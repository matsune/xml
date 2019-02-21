package xml

import (
	"fmt"
	"reflect"
	"testing"
)

func newRune(r rune) *rune { return &r }

func TestParser_parseXmlDecl(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		want    *XMLDecl
		wantErr bool
	}{
		{
			name:    "not <?xml",
			str:     `<xml version="1.0" standalone="no" ?>`,
			wantErr: true,
		},
		{
			name:    "error version",
			str:     `<?xml version=1.0 standalone="no" ?>`,
			wantErr: true,
		},
		{
			name:    "error encoding",
			str:     `<?xml version="1.0" encoding= ?>`,
			wantErr: true,
		},
		{
			name:    "error standalone",
			str:     `<?xml version="1.0" standalone= ?>`,
			wantErr: true,
		},
		{
			name:    "error close",
			str:     `<?xml version="1.0"  >`,
			wantErr: true,
		},
		{
			str: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`,
			want: &XMLDecl{
				Version:    "1.0",
				Encoding:   "UTF-8",
				Standalone: true,
			},
		},
		{
			str: `<?xml version="1.1"  standalone="no"    ?>`,
			want: &XMLDecl{
				Version:    "1.1",
				Encoding:   "",
				Standalone: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.str)
			got, err := p.parseXmlDecl()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseXmlDecl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.parseXmlDecl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseVersion(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantVer string
		wantErr bool
	}{
		{
			name:    "not starts with spaces",
			source:  `version="1.0"`,
			wantErr: true,
		},
		{
			name:    "not starts with version",
			source:  ` ver="1.0"`,
			wantErr: true,
		},
		{
			name:    "not equal",
			source:  ` version:"1.0"`,
			wantErr: true,
		},
		{
			name:    "no quote",
			source:  ` version=1.0`,
			wantErr: true,
		},
		{
			name:    "error while parsing version num",
			source:  ` version=""`,
			wantErr: true,
		},
		{
			name:    "different quotes",
			source:  ` version="1.0'`,
			wantErr: true,
		},
		{
			source:  ` version="1.0" `,
			wantVer: "1.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			gotVer, err := p.parseVersion()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVer != tt.wantVer {
				t.Errorf("Parser.parseVersion() = %v, want %v", gotVer, tt.wantVer)
			}
		})
	}
}

func TestParser_parseName(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    string
		wantErr bool
	}{
		{
			name:    "starts with null char",
			source:  fmt.Sprintf("%caaa", 0),
			wantErr: true,
		},
		{
			source: ":abc",
			want:   ":abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parseName()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Parser.parseName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseSystemLiteral(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    string
		wantErr bool
	}{
		{
			name:    "not starts with quote",
			source:  `aaa`,
			wantErr: true,
		},
		{
			name:    "different quotes",
			source:  `'a"`,
			wantErr: true,
		},
		{
			name:   "empty literal",
			source: `''`,
			want:   "",
		},
		{
			source: `"aaa"`,
			want:   "aaa",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parseSystemLiteral()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseSystemLiteral() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Parser.parseSystemLiteral() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parsePubidLiteral(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    string
		wantErr bool
	}{
		{
			name:    "not starts with quote",
			source:  `aaa`,
			wantErr: true,
		},
		{
			name:    "different quotes",
			source:  `'a"`,
			wantErr: true,
		},
		{
			name:   "empty literal",
			source: `''`,
			want:   "",
		},
		{
			source: `"aaa"`,
			want:   "aaa",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parsePubidLiteral()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parsePubidLiteral() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Parser.parsePubidLiteral() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseComment(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    Comment
		wantErr bool
	}{
		{
			name:    "not starts with <!--",
			source:  "<-- aa -->",
			wantErr: true,
		},
		{
			name:    "not end with -->",
			source:  "<!-- aa --",
			wantErr: true,
		},
		{
			name:    "contains not char",
			source:  fmt.Sprintf("<!-- %c -->", 0x0),
			wantErr: true,
		},
		{
			source: "<!-- this is comment-->",
			want:   Comment(" this is comment"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parseComment()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Parser.parseComment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseMarkup(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    Markup
		wantErr bool
	}{
		{
			name:    "unknown markup",
			source:  "",
			wantErr: true,
		},
		{
			name:   "element",
			source: "<!ELEMENT student (id|(a,b)?)>",
			want: &Element{
				Name: "student",
				ContentSpec: &Children{
					ChoiceSeq: &Choice{
						CPs: []CP{
							CP{
								Name: "id",
							},
							CP{
								ChoiceSeq: &Seq{
									CPs: []CP{
										CP{
											Name: "a",
										},
										CP{
											Name: "b",
										},
									},
								},
								Suffix: newRune('?'),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parseMarkup()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseMarkup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.parseMarkup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseStandalone(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    bool
		wantErr bool
	}{
		{
			name:    "not start with spaces",
			source:  `standalone='yes'`,
			wantErr: true,
		},
		{
			name:    "no standalone",
			source:  ` stand='yes'`,
			wantErr: true,
		},
		{
			name:    "error parse =",
			source:  ` standalone:'yes'`,
			wantErr: true,
		},
		{
			name:    "no quote",
			source:  ` standalone=yes`,
			wantErr: true,
		},
		{
			name:    "invalid bool value",
			source:  ` standalone='true'`,
			wantErr: true,
		},
		{
			name:    "difference quotes",
			source:  ` standalone="yes'`,
			wantErr: true,
		},
		{
			source: ` standalone="yes"`,
			want:   true,
		},
		{
			source: ` standalone='no'`,
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parseStandalone()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseStandalone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Parser.parseStandalone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseElement(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    *Element
		wantErr bool
	}{
		{
			name:    "invalid head",
			source:  `<ELEMENT student (id)>`,
			wantErr: true,
		},
		{
			name:    "no space",
			source:  `<!ELEMENTstudent (id)>`,
			wantErr: true,
		},
		{
			name:    "name error",
			source:  `<!ELEMENT (id)>`,
			wantErr: true,
		},
		{
			name:    "no space after name name",
			source:  `<!ELEMENT student(id)>`,
			wantErr: true,
		},
		{
			name:    "invalid content spec",
			source:  `<!ELEMENT student ()>`,
			wantErr: true,
		},
		{
			name:    "not closed",
			source:  `<!ELEMENT student (id)`,
			wantErr: true,
		},
		{
			name:   "simple element",
			source: `<!ELEMENT student (id)>`,
			want: &Element{
				Name: "student",
				ContentSpec: &Children{
					ChoiceSeq: &Choice{
						CPs: []CP{
							CP{
								Name: "id",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parseElement()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseElement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.parseElement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseContentSpec(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    ContentSpec
		wantErr bool
	}{
		{
			name:   "parse EMPTY",
			source: "EMPTY",
			want:   &EMPTY{},
		},
		{
			name:   "parse ANY",
			source: "ANY",
			want:   &ANY{},
		},
		{
			name:   "parse mixed",
			source: "(#PCDATA)",
			want:   &Mixed{},
		},
		{
			name:   "parse children",
			source: "(id)",
			want: &Children{
				ChoiceSeq: &Choice{
					CPs: []CP{
						CP{
							Name: "id",
						},
					},
				},
			},
		},
		{
			name:    "error",
			source:  "(id",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parseContentSpec()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseContentSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.parseContentSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseChildren(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    *Children
		wantErr bool
	}{
		{
			name:    "parse error",
			source:  "id|name)",
			wantErr: true,
		},
		{
			name:   "choice children",
			source: "(id|name)",
			want: &Children{
				ChoiceSeq: &Choice{
					CPs: []CP{
						CP{
							Name: "id",
						},
						CP{
							Name: "name",
						},
					},
				},
			},
		},
		{
			name:   "choice children with suffix",
			source: "(id|name)+",
			want: &Children{
				ChoiceSeq: &Choice{
					CPs: []CP{
						CP{
							Name: "id",
						},
						CP{
							Name: "name",
						},
					},
				},
				Suffix: newRune('+'),
			},
		},
		{
			name:   "seq children with suffix",
			source: "(id,name)+",
			want: &Children{
				ChoiceSeq: &Seq{
					CPs: []CP{
						CP{
							Name: "id",
						},
						CP{
							Name: "name",
						},
					},
				},
				Suffix: newRune('+'),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parseChildren()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseChildren() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.parseChildren() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseCP(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    *CP
		wantErr bool
	}{
		{
			name:    "error while parsing seq",
			source:  "(surname,)",
			wantErr: true,
		},
		{
			name:   "nested",
			source: "(surname,(origin|sex)?)",
			want: &CP{
				ChoiceSeq: &Seq{
					CPs: []CP{
						CP{
							Name: "surname",
						},
						CP{
							ChoiceSeq: &Choice{
								CPs: []CP{
									CP{
										Name: "origin",
									},
									CP{
										Name: "sex",
									},
								},
							},
							Suffix: newRune('?'),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parseCP()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseCP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.parseCP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseChoice(t *testing.T) {
	type fields struct {
		Scanner *Scanner
	}
	tests := []struct {
		name    string
		source  string
		want    *Choice
		wantErr bool
	}{
		{
			name:    "not starts with (",
			source:  `surname|firstname)`,
			wantErr: true,
		},
		{
			name:    "error no cp",
			source:  `()`,
			wantErr: true,
		},
		{
			name: "error parsing cp",
			source: `(surname|
				)`,
			wantErr: true,
		},
		{
			name:    "not closed",
			source:  `(surname|firstname`,
			wantErr: true,
		},
		{
			source: `(surname|firstname)`,
			want: &Choice{
				CPs: []CP{
					CP{
						Name: "surname",
					},
					CP{
						Name: "firstname",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parseChoice()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseChoice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.parseChoice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseSeq(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    *Seq
		wantErr bool
	}{
		{
			name:    "not starts with (",
			source:  `surname,firstname*,dob?,(origin|sex)?)`,
			wantErr: true,
		},
		{
			name:    "error no cp",
			source:  `()`,
			wantErr: true,
		},
		{
			name: "error parsing cp",
			source: `(surname,
				)`,
			wantErr: true,
		},
		{
			name:    "not closed",
			source:  `(surname,firstname*`,
			wantErr: true,
		},
		{
			source: `(surname,firstname*)`,
			want: &Seq{
				CPs: []CP{
					CP{
						Name: "surname",
					},
					CP{
						Name:   "firstname",
						Suffix: newRune('*'),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parseSeq()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseSeq() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.parseSeq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseMixed(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    *Mixed
		wantErr bool
	}{
		{
			name:    "not starts with (",
			source:  "#PCDATA)*",
			wantErr: true,
		},
		{
			name:    "no #PCDATA",
			source:  "()*",
			wantErr: true,
		},
		{
			name:    "not | between #PCDATA and )",
			source:  "(#PCDATA  a)*",
			wantErr: true,
		},
		{
			name:    "no name after #PCDATA|",
			source:  "(#PCDATA|)*",
			wantErr: true,
		},
		{
			name:    "not closed",
			source:  "(#PCDATA|a",
			wantErr: true,
		},
		{
			name:    "error if has names but no *",
			source:  "(#PCDATA|a)",
			wantErr: true,
		},
		{
			source: "(#PCDATA|a  |  b)*",
			want: &Mixed{
				Names: []string{"a", "b"},
			},
		},
		{
			source: "(  #PCDATA  )",
			want:   &Mixed{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parseMixed()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseMixed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.parseMixed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseAttlist(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    *Attlist
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parseAttlist()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseAttlist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.parseAttlist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseEntityReference(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    EntityRef
		wantErr bool
	}{
		{
			name:    "not starts with &",
			source:  `name;`,
			wantErr: true,
		},
		{
			name:    "error parse name",
			source:  `&;`,
			wantErr: true,
		},
		{
			name:    "not closed ;",
			source:  `&name`,
			wantErr: true,
		},
		{
			source: `&name;`,
			want:   EntityRef("name"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parseEntityReference()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseEntityReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Parser.parseEntityReference() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parsePEReference(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    PERef
		wantErr bool
	}{
		{
			name:    "not starts with %",
			source:  `name;`,
			wantErr: true,
		},
		{
			name:    "error parse name",
			source:  `%;`,
			wantErr: true,
		},
		{
			name:    "not closed ;",
			source:  `%name`,
			wantErr: true,
		},
		{
			source: `%name;`,
			want:   PERef("name"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parsePEReference()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parsePEReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Parser.parsePEReference() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseExternalID(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    *ExternalID
		wantErr bool
	}{
		{
			name:    "not starts with SYSTEM or PUBLIC",
			source:  `SYS "aaa"`,
			wantErr: true,
		},
		{
			name:    "no space",
			source:  `SYSTEM"aaa"`,
			wantErr: true,
		},
		{
			name:    "pubid literal error",
			source:  `PUBLIC aa"`,
			wantErr: true,
		},
		{
			name:    "no space after pubid literal",
			source:  `PUBLIC "aa"`,
			wantErr: true,
		},
		{
			name:    "system literal error",
			source:  `SYSTEM aa"`,
			wantErr: true,
		},
		{
			name:   "public no error",
			source: `PUBLIC "pub" "sys"`,
			want: &ExternalID{
				Identifier: ExtPublic,
				Pubid:      "pub",
				System:     "sys",
			},
		},
		{
			name:   "system no error",
			source: `SYSTEM "sys"`,
			want: &ExternalID{
				Identifier: ExtSystem,
				System:     "sys",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parseExternalID()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseExternalID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.parseExternalID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseEncoding(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    string
		wantErr bool
	}{
		{
			name:    "not starts with spaces",
			source:  `encoding="UTF-8"`,
			wantErr: true,
		},
		{
			name:    "typo encoding",
			source:  ` enco="UTF-8"`,
			wantErr: true,
		},
		{
			name:    "not equal",
			source:  ` encoding:"UTF-8"`,
			wantErr: true,
		},
		{
			name:    "no quote",
			source:  ` encoding=UTF-8`,
			wantErr: true,
		},
		{
			name:    "error while parsing encoding name",
			source:  ` encoding="あ" `,
			wantErr: true,
		},
		{
			name:    "different quote",
			source:  ` encoding="UTF-8' `,
			wantErr: true,
		},
		{
			source: ` encoding="UTF-8" `,
			want:   "UTF-8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parseEncoding()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseEncoding() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Parser.parseEncoding() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseEncName(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    string
		wantErr bool
	}{
		{
			name:    "not starts with alphabet",
			source:  "8UTF",
			wantErr: true,
		},
		{
			source: "UTF-8",
			want:   "UTF-8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.source)
			got, err := p.parseEncName()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseEncName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Parser.parseEncName() = %v, want %v", got, tt.want)
			}
		})
	}
}
