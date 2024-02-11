// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.543
package templates

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

import (
	"github.com/datasektionen/pls4/models"
)

func options(permissions []models.SystemPermissions) bool {
	for _, p := range permissions {
		if p.MayEdit {
			return true
		}
	}
	return false
}

func Permissions(permissions []models.SystemPermissions) templ.Component {
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
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<section class=\"grid grid-cols-[auto_1fr{{ if $options }}_1fr{{ end }}] gap-x-6 gap-y-2 items-center p-3\" id=\"members\" hx-swap=\"outerHTML\" hx-target=\"this\"><p class=\"font-bold\">System</p><p class=\"font-bold\">Permission</p>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if options(permissions) {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<p class=\"font-bold\">Options</p>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		for _, permission := range permissions {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<hr class=\"col-span-full\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			for _, perm := range permission.Permissions {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<p>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var2 string
				templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(permission.System)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `permissions.templ`, Line: 30, Col: 37}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</p><p>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var3 string
				templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(perm)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `permissions.templ`, Line: 31, Col: 24}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</p>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				if permission.MayEdit {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<p class=\"text-red-800\">Remove</p>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				} else if options(permissions) {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<p></p>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</section>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
