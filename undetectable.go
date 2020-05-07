package chromedp

import (
	"context"

	"github.com/chromedp/cdproto/page"
)

/*
			chromedp.Flag("enable-automation", false),
			chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Safari/537.36"),

	// chromedp.Navigate(`https://intoli.com/blog/making-chrome-headless-undetectable/chrome-headless-test.html`),
		// chromedp.Navigate(`https://intoli.com/blog/not-possible-to-block-chrome-headless/chrome-headless-test.html`),
		chromedp.EmulateViewport(1920, 3000),
		chromedp.Navigate(`https://antoinevastel.com/bots/`),
		chromedp.Sleep(2*time.Second),
		chromedp.CaptureScreenshot(&buf),
*/

type undetectableOptions struct {
	bypassIframeTest bool // default true
}

func newUndetectableOptions() *undetectableOptions {
	return &undetectableOptions{
		bypassIframeTest: true,
	}
}

type UndetectableOption = func(*undetectableOptions)

func BypassIframeTest(bypassIframeTest bool) UndetectableOption {
	return func(opts *undetectableOptions) {
		opts.bypassIframeTest = bypassIframeTest
	}
}

func Undetectable(opts ...UndetectableOption) Action {
	script := `
	(function (window, navigator) {
		delete navigator.__proto__.webdriver;
	
		// The method below cant bypass "'webdriver' in navigator", so we remove it
		// // Pass the Webdriver Test.
		// const originHasOwnProperty = navigator.hasOwnProperty;
		// navigator.hasOwnProperty = (property) => (
		// 	property === 'webdriver' ? false : originHasOwnProperty(property)
		// );
	
		// The method below cant bypass "'webdriver' in navigator", so we remove it
		// Object.defineProperty(navigator, 'webdriver', {
		//   get: () => undefined,
		// });

		// This value was be null when using chrome-driver
		Object.defineProperty(navigator, 'doNotTrack', {
			get: () => '1',
		});
	
		// Pass the Plugins Length Test.
		// Overwrite the plugins property to use a custom getter.
		udNewPlugin = function(prot, mimes) {
			prot.__proto__ = Plugin.prototype
			p = (function () {
				var res = {};
				for (let i = 0; i < mimes.length; i++) {
					Object.defineProperty(res, i+"", {
						configurable: true,
						enumerable: true,
						get: function () {
							mimes[i].__proto__ = MimeType.prototype
							return mimes[i];
						}
					});
				}
				Object.defineProperty(res, "length", {
					get: function () {
						return mimes.length
					}
				});
				Object.setPrototypeOf(res, prot);
				return res;
			})();
			// doesn't need
			// for (let i = 0; i < mimes.length; i++) {
			// 	mimes[i].enabledPlugin = p
			// }
			return p
		}
		udMimes = [
			{
				type: "application/pdf",
				suffixes: "pdf",
				description: "",
			},
			{
				type: "application/x-google-chrome-pdf",
				suffixes: "pdf",
				description: "Portable Document Format",
			},
			{
				type: "application/x-nacl",
				suffixes: "",
				description: "Native Client Executable",
			},
			{
				type: "application/x-pnacl",
				suffixes: "",
				description: "Portable Native Client Executable",
			}
		]
		udMimes.__proto__ = MimeTypeArray.prototype
		udPlugins = [
			udNewPlugin({
				name: "Chrome PDF Plugin",
				filename: "internal-pdf-viewer",
				description: "Portable Document Format",
			},[udMimes[1]]),
			udNewPlugin({
				name: "Chrome PDF Viewer",
				filename: "mhjfbmdgcfjbbpaeojofohoefgiehjai",
				description: "",
			},[udMimes[0]]),
			udNewPlugin({
				name: "Native Client",
				filename: "internal-nacl-plugin",
				description: "",
			},[udMimes[2],udMimes[3]]),
		]
		udPlugins.__proto__ = PluginArray.prototype
		Object.defineProperty(navigator, 'mimeTypes', {
			get: () => {
				return udMimes;
			},
		});
		Object.defineProperty(navigator, 'plugins', {
			get: () => {
				return udPlugins;
			},
		});

		// Pass the Languages Test.
		// Overwrite the languages property to use a custom getter.
		Object.defineProperty(navigator, 'languages', {
			get: () => ['en-US', 'en'],
		});
	
		// Pass the Chrome Test.
		// We can mock this in as much depth as we need for the test.
		window.chrome = {
			runtime: {},
		};
	
		// Pass the Permissions Test.
		const originalQuery = window.navigator.permissions.query;
		return window.navigator.permissions.query = (parameters) => (
			parameters.name === 'notifications' ?
				Promise.resolve({ state: Notification.permission }) :
				originalQuery(parameters)
		);
	})(window, navigator, window.navigator);
	
	//
	// Bypass the WebGL test.
	//
	const originalGetParameter = WebGLRenderingContext.getParameter;
	WebGLRenderingContext.prototype.getParameter = function (parameter) {
		// UNMASKED_VENDOR_WEBGL
		if (parameter === 37445) {
			return 'Intel Open Source Technology Center';
		}
		// UNMASKED_RENDERER_WEBGL
		if (parameter === 37446) {
			return 'Mesa DRI Intel(R) Ivybridge Mobile ';
		}
	
		return originalGetParameter(parameter);
	};
	
	
	//
	// Bypass the Broken Image Test.
	//
	
	['height', 'width'].forEach(property => {
		// Store the existing descriptor.
		const imageDescriptor = Object.getOwnPropertyDescriptor(HTMLImageElement.prototype, property);
	
		// Redefine the property with a patched descriptor.
		Object.defineProperty(HTMLImageElement.prototype, property, {
			...imageDescriptor,
			get: function () {
				// Return an arbitrary non-zero dimension if the image failed to load.
				if (this.complete && this.naturalHeight == 0) {
					return 20;
				}
				// Otherwise, return the actual dimension.
				return imageDescriptor.get.apply(this);
			},
		});
	});
	
	
	//
	// Bypass the Retina/HiDPI Hairline Feature Test.
	//
	
	// Store the existing descriptor.
	const elementDescriptor = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetHeight');
	
	// Redefine the property with a patched descriptor.
	Object.defineProperty(HTMLDivElement.prototype, 'offsetHeight', {
		...elementDescriptor,
		get: function () {
			if (this.id === 'modernizr') {
				return 1;
			}
			return elementDescriptor.get.apply(this);
		},
	});

	  `
	options := newUndetectableOptions()
	for _, opt := range opts {
		opt(options)
	}

	if options.bypassIframeTest {
		script = script + `
// Pass the iframe Test
	Object.defineProperty(HTMLIFrameElement.prototype, 'contentWindow', {
		get: function() {
		  return window;
		}
	});
`
	}

	return ActionFunc(func(ctx context.Context) error {
		if _, err := page.AddScriptToEvaluateOnNewDocument(script).Do(ctx); err != nil {
			return err
		}
		return nil
	})
}
