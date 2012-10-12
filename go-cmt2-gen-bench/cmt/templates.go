package cmt

import (
	"text/template"
)

var hdr_tmpl = template.Must(template.New("hdr").Parse(
	`/* -*- c++ -*- */
#ifndef LIB_{{.Name}}_H
#define LIB_{{.Name}}_H 1

{{with .Uses}}{{range .}}#include "{{.Name}}/Lib{{.Name}}.h"
{{end}}{{end}}

#ifdef _MSC_VER
# define API_EXPORT __declspec( dllexport )
#else
#if __GNUC__ >= 4
# define API_EXPORT __attribute__((visibility("default")))
#else
# define API_EXPORT
#endif
#endif

class API_EXPORT C{{.Name}}
{
public:
   C{{.Name}}();
   ~C{{.Name}}();
   void f();
private:
{{with .Uses}}{{range .}} C{{.Name}} m_{{.Name}};
{{end}}{{end}}
};
#endif /* !LIB_{{.Name}}_H */
/* EOF */
`))

var cxx_tmpl = template.Must(template.New("cxx").Parse(
`// Lib{{.Name}}.cxx
#include <iostream>
#include "{{.Name}}/Lib{{.Name}}.h"

C{{.Name}}::C{{.Name}}()
{
   std::cout << ":: c-tor C{{.Name}}\n";
}

C{{.Name}}::~C{{.Name}}()
{
   std::cout << ":: d-tor C{{.Name}}\n";
}

void
C{{.Name}}::f()
{
   std::cout << ":: C{{.Name}}.f\n";
   {{with .Uses}}{{range .}}m_{{.Name}}.f();
   {{end}}{{end}}
}
`))

// EOF
