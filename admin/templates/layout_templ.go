// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.543
package templates

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

func methone(userID, loginURL string) templ.ComponentScript {
	return templ.ComponentScript{
		Name: `__templ_methone_003b`,
		Function: `function __templ_methone_003b(userID, loginURL){window.methone_conf = {
		system_name: "pls",
		color_scheme: "purple",
		login_text: userID ? userID : "Login",
		login_href: userID ? "/logout" : loginURL
			+ "/login?callback="
			+ encodeURIComponent(window.location.origin + "/login?code="),
	};
}`,
		Call:       templ.SafeScript(`__templ_methone_003b`, userID, loginURL),
		CallInline: templ.SafeScriptInline(`__templ_methone_003b`, userID, loginURL),
	}
}

func Layout(userID, loginURL string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
		if !templ_7745c5c3_IsBuffer {
			templ_7745c5c3_Buffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<!doctype html><html lang=\"sv\"><head><meta charset=\"UTF-8\"><meta http-equiv=\"X-UA-Compatible\" content=\"IE=edge\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"><title>pls - fourth* time's the charm</title>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = methone(userID, loginURL).Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<script defer src=\"https://methone.datasektionen.se/bar.js\"></script><script src=\"https://cdn.tailwindcss.com\"></script><script src=\"https://unpkg.com/htmx.org@1.9.5\"></script><style>\n\t\t\t\t@import url(https://fonts.googleapis.com/css?family=Lato:400,300,700,400italic,700italic,900);\n\t\t\t\t@import url(https://use.fontawesome.com/releases/v6.4.2/css/all.css);\n\t\t\t\tbody {\n\t\t\t\t\tfont-family: Lato;\n\t\t\t\t}\n\t\t</style></head><body><div class=\"h-[50px]\" id=\"methone-container-replace\"></div><main class=\"p-4 max-w-screen-lg mx-auto md:mt-24\" hx-boost=\"true\" hx-target=\"this\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ_7745c5c3_Var1.Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</main></body></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
