// Copyright (c) 2026 Michael Lechner. All rights reserved.
package executor

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	v8 "rogchap.com/v8go"
)

// InjectPolyfills adds standard APIs to the V8 context that are missing in raw V8.
func InjectPolyfills(iso *v8.Isolate, ctx *v8.Context) error {
	global := ctx.Global()

	// 1. performance.now()
	start := time.Now()
	perfObj := v8.NewObjectTemplate(iso)
	nowFn := v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		ms := float64(time.Since(start).Nanoseconds()) / 1e6
		val, _ := v8.NewValue(iso, ms)
		return val
	})
	_ = perfObj.Set("now", nowFn)
	perfInst, _ := perfObj.NewInstance(ctx)
	_ = global.Set("performance", perfInst)

	// 2. Base64 (atob / btoa)
	atobFn := v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		if len(info.Args()) < 1 {
			return v8.Undefined(iso)
		}
		data, err := base64.StdEncoding.DecodeString(info.Args()[0].String())
		if err != nil {
			return v8.Undefined(iso)
		}
		val, _ := v8.NewValue(iso, string(data))
		return val
	})
	btoaFn := v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		if len(info.Args()) < 1 {
			return v8.Undefined(iso)
		}
		str := base64.StdEncoding.EncodeToString([]byte(info.Args()[0].String()))
		val, _ := v8.NewValue(iso, str)
		return val
	})
	_ = global.Set("atob", atobFn.GetFunction(ctx))
	_ = global.Set("btoa", btoaFn.GetFunction(ctx))

	// 3. Internal helper for crypto.getRandomValues (returning base64 for reliable bridge)
	getRandomB64 := v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		if len(info.Args()) < 1 {
			return v8.Undefined(iso)
		}
		n := int(info.Args()[0].Integer())
		if n <= 0 {
			return v8.Undefined(iso)
		}
		b := make([]byte, n)
		_, _ = rand.Read(b)
		val, _ := v8.NewValue(iso, base64.StdEncoding.EncodeToString(b))
		return val
	})
	_ = global.Set("__get_random_b64", getRandomB64.GetFunction(ctx))

	// 4. UTF-8 Bridge
	encodeUtf8 := v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		if len(info.Args()) < 1 {
			return v8.Undefined(iso)
		}
		str := info.Args()[0].String()
		bytes := []byte(str)

		// Create a JS array of numbers
		jsArr, _ := v8.NewObjectTemplate(iso).NewInstance(ctx)
		for i, b := range bytes {
			val, _ := v8.NewValue(iso, int32(b))
			_ = jsArr.SetIdx(uint32(i), val)
		}
		_ = jsArr.Set("length", len(bytes))
		return jsArr.Value
	})
	decodeUtf8 := v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		if len(info.Args()) < 1 {
			return v8.Undefined(iso)
		}
		jsArr := info.Args()[0].Object()
		lenVal, _ := jsArr.Get("length")
		length := int(lenVal.Integer())

		bytes := make([]byte, length)
		for i := 0; i < length; i++ {
			val, _ := jsArr.GetIdx(uint32(i))
			bytes[i] = byte(val.Integer())
		}

		val, _ := v8.NewValue(iso, string(bytes))
		return val
	})
	_ = global.Set("__encode_utf8", encodeUtf8.GetFunction(ctx))
	_ = global.Set("__decode_utf8", decodeUtf8.GetFunction(ctx))

	// 5. JS-side Polyfills (Crypto, TextEncoder, Buffer)
	_, err := ctx.RunScript(`
		(function() {
			// Helper to convert b64 to Uint8Array
			function b64ToUint8(b64) {
				const s = atob(b64);
				const b = new Uint8Array(s.length);
				for (let i = 0; i < s.length; i++) b[i] = s.charCodeAt(i);
				return b;
			}

			// Crypto
			globalThis.crypto = {
				getRandomValues: function(array) {
					const bytes = b64ToUint8(__get_random_b64(array.byteLength));
					const view = new Uint8Array(array.buffer, array.byteOffset, array.byteLength);
					for (let i = 0; i < bytes.length; i++) view[i] = bytes[i];
					return array;
				}
			};

			// TextEncoder / TextDecoder using the Go bridge
			globalThis.TextEncoder = class {
				encode(str) {
					const arr = __encode_utf8(str);
					return new Uint8Array(Object.values(arr));
				}
			};
			globalThis.TextDecoder = class {
				decode(buf) {
					const arr = Array.from(new Uint8Array(buf));
					return __decode_utf8(arr);
				}
			};

			// Minimal Buffer
			globalThis.Buffer = {
				from: function(data, encoding) {
					if (typeof data === 'string') {
						if (encoding === 'base64') return b64ToUint8(data);
						return new TextEncoder().encode(data);
					}
					return new Uint8Array(data);
				},
				alloc: (size) => new Uint8Array(size)
			};
		})();
	`, "polyfills.js")

	return err
}
