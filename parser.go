package xml

import (
	"errors"
	"fmt"
	"strings"
)

type parser struct {
	*scanner
	parsing string
}

func (p *parser) error(err error) *XMLError {
	return newErr(p.parsing, err, p.pos())
}

func (p *parser) setParsing(new string) func() {
	old := p.parsing
	p.parsing = new
	return func() {
		p.parsing = old
	}
}

func (p *parser) Test(r rune) bool {
	return p.Get() == r
}

func (p *parser) Must(r rune) error {
	if !p.Test(r) {
		return p.error(fmt.Errorf("expected %q", r))
	}
	p.Step()
	return nil
}

func (p *parser) Tests(str string) bool {
	i := int(p.cursor)
	e := i + len([]rune(str))
	if len(p.source) < e {
		return false
	}
	return string(p.source[i:e]) == str
}

func (p *parser) Musts(str string) error {
	if !p.Tests(str) {
		return p.error(fmt.Errorf("expected %q", str))
	}
	p.StepN(len(str))
	return nil
}

/// EBNF for XML 1.0
/// http://www.jelks.nu/XML/xmlebnf.html#NT-VersionInfo

// - Document

// document ::= prolog element Misc*
func (p *parser) parse() (*XML, error) {
	var x XML
	var err error
	x.Prolog, err = p.parseProlog()
	if err != nil {
		return nil, err
	}
	x.Element, err = p.parseElement()
	if err != nil {
		return nil, err
	}
	for {
		cur := p.cursor
		var misc Misc
		if misc, err = p.parseMisc(); err != nil {
			p.cursor = cur
			err = nil
			break
		}
		if misc != nil {
			x.Misc = append(x.Misc, misc)
		}
	}
	return &x, nil
}

/// - Prolog

// prolog ::= XMLDecl? Misc* (doctypedecl Misc*)?
func (p *parser) parseProlog() (*Prolog, error) {
	pro := Prolog{}

	p.skipSpace()

	if p.Tests("<?xml") {
		xmlDecl, err := p.parseXmlDecl()
		if err != nil {
			return nil, err
		}
		pro.XMLDecl = xmlDecl
	}

	for {
		cur := p.cursor

		if misc, err := p.parseMisc(); err != nil {
			p.cursor = cur
			err = nil
			break
		} else if misc != nil {
			pro.Misc1 = append(pro.Misc1, misc)
		}
	}

	if p.Tests(`<!DOCTYPE`) {
		doc, err := p.parseDoctype()
		if err != nil {
			return nil, err
		}
		pro.DOCType = doc

		for {
			cur := p.cursor

			if misc, err := p.parseMisc(); err != nil {
				p.cursor = cur
				err = nil
				break
			} else if misc != nil {
				pro.Misc2 = append(pro.Misc2, misc)
			}
		}
	}
	return &pro, nil
}

// XMLDecl ::= '<?xml' VersionInfo EncodingDecl? SDDecl? S? '?>'
func (p *parser) parseXmlDecl() (*XMLDecl, error) {
	defer p.setParsing("XML Declaration")()

	if err := p.Musts("<?xml"); err != nil {
		return nil, err
	}
	x := XMLDecl{}

	ver, err := p.parseVersion()
	if err != nil {
		return nil, err
	}
	x.Version = ver

	// keep cursor at this time
	cur := p.cursor

	p.skipSpace()

	if p.Tests("encoding") {
		// reset cursor before skipping spaces
		p.cursor = cur

		enc, err := p.parseEncoding()
		if err != nil {
			return nil, err
		}
		x.Encoding = enc
	} else {
		p.cursor = cur
	}

	cur = p.cursor

	p.skipSpace()

	if p.Tests("standalone") {
		p.cursor = cur

		std, err := p.parseStandalone()
		if err != nil {
			return nil, err
		}
		x.Standalone = std
	} else {
		p.cursor = cur
	}

	p.skipSpace()
	if err := p.Musts("?>"); err != nil {
		return nil, err
	}

	return &x, nil
}

// VersionInfo ::= S 'version' Eq (' VersionNum ' | " VersionNum ")
func (p *parser) parseVersion() (string, error) {
	var err error
	if err = p.parseSpace(); err != nil {
		return "", err
	}

	if err = p.Musts("version"); err != nil {
		return "", err
	}
	if err = p.parseEq(); err != nil {
		return "", err
	}

	var quote rune
	quote, err = p.parseQuote()
	if err != nil {
		return "", err
	}

	var ver string
	ver, err = p.parseVersionNum()
	if err != nil {
		return "", err
	}

	if err = p.Must(quote); err != nil {
		return "", err
	}
	return ver, nil
}

// Eq ::= S? '=' S?
func (p *parser) parseEq() error {
	p.skipSpace()
	if err := p.Must('='); err != nil {
		return err
	}
	p.skipSpace()
	return nil
}

// VersionNum ::= ([a-zA-Z0-9_.:] | '-')+
func (p *parser) parseVersionNum() (string, error) {
	isVerChar := func() (rune, bool) {
		r := p.Get()
		if isNum(r) || isAlpha(r) || p.Test('_') || p.Test('.') || p.Test(':') || p.Test('-') {
			return r, true
		} else {
			return r, false
		}
	}

	var str string
	r, ok := isVerChar()
	if !ok {
		return "", p.error(fmt.Errorf("unexpected version number character '%c'", r))
	}
	str += string(r)
	p.Step()

	for {
		r, ok := isVerChar()
		if !ok {
			break
		}
		str += string(r)
		p.Step()
	}

	return str, nil
}

// Misc ::= Comment | PI | S
func (p *parser) parseMisc() (Misc, error) {
	if p.Tests(`<!--`) {
		c, err := p.parseComment()
		if err != nil {
			return nil, err
		}
		return c, nil
	} else if p.Tests(`<?`) {
		pi, err := p.parsePI()
		if err != nil {
			return nil, err
		}
		return pi, err
	} else if isSpace(p.Get()) {
		p.skipSpace()
		return nil, nil
	} else {
		return nil, p.error(fmt.Errorf("unknown misc type"))
	}
}

/// - White Space

// S ::= (#x20 | #x9 | #xD | #xA)+
func (p *parser) parseSpace() error {
	if !isSpace(p.Get()) {
		return p.error(fmt.Errorf("expected space"))
	}
	p.skipSpace()
	return nil
}

// Skip spaces until reaches not space
func (p *parser) skipSpace() {
	for isSpace(p.Get()) {
		p.Step()
	}
}

/// - Names and Tokens

// NameChar ::= Letter | Digit | '.' | '-' | '_' | ':' | CombiningChar |  Extender
func (p *parser) isNameChar() bool {
	return isLetter(p.Get()) || isDigit(p.Get()) || p.Test('.') || p.Test('-') || p.Test('_') || p.Test(':') || isCombining(p.Get()) || isExtender(p.Get())
}

// Name ::= (Letter | '_' | ':') (NameChar)*
func (p *parser) parseName() (string, error) {
	var n string
	if isLetter(p.Get()) || p.Test('_') || p.Test(':') {
		n += string(p.Get())
		p.Step()
	} else {
		return "", p.error(errors.New("invalid letter for name"))
	}
	for p.isNameChar() {
		n += string(p.Get())
		p.Step()
	}
	return n, nil
}

/// - Literals

// EntityValue ::= '"' ([^%&"] | PEReference | Reference)* '"' |  "'" ([^%&'] |  PEReference |  Reference)* "'"
func (p *parser) parseEntityValue() (EntityValue, error) {
	var quote rune
	var err error
	if quote, err = p.parseQuote(); err != nil {
		return nil, err
	}

	res := EntityValue{}

	var str string
	for {
		if p.Test(quote) || p.isEnd() {
			break
		}

		cur := p.cursor

		if p.Test('&') {
			if len(str) > 0 {
				res = append(res, str)
				str = ""
			}

			// try EntityRef
			var eRef *EntityRef
			eRef, err = p.parseEntityRef()
			if err != nil {
				p.cursor = cur
				// try CharRef
				var cRef *CharRef
				cRef, err = p.parseCharRef()
				if err != nil {
					return nil, p.error(errors.New("error AttValue"))
				}
				res = append(res, cRef)
			} else {
				res = append(res, eRef)
			}
		} else if p.Test('%') {
			if len(str) > 0 {
				res = append(res, str)
				str = ""
			}

			var pRef *PERef
			if pRef, err = p.parsePERef(); err != nil {
				return nil, err
			}
			res = append(res, pRef)
		} else {
			str += string(p.Get())
			p.Step()
		}
	}

	if len(str) > 0 {
		res = append(res, str)
		str = ""
	}

	if err = p.Must(quote); err != nil {
		return nil, err
	}

	return res, nil
}

// AttValue ::= '"' ([^<&"] | Reference)* '"' |  "'" ([^<&'] | Reference)* "'"
func (p *parser) parseAttValue() (AttValue, error) {
	var quote rune
	var err error
	if quote, err = p.parseQuote(); err != nil {
		return nil, err
	}

	res := AttValue{}

	var str string
	for {
		if p.Test('<') {
			return nil, p.error(fmt.Errorf("unexpected '<'"))
		}
		if p.Test(quote) || p.isEnd() {
			break
		}

		if p.Test('&') {
			if len(str) > 0 {
				res = append(res, str)
				str = ""
			}

			var ref Ref
			if ref, err = p.parseReference(); err != nil {
				return nil, err
			}
			res = append(res, ref)
		} else {
			str += string(p.Get())
			p.Step()
		}
	}

	if len(str) > 0 {
		res = append(res, str)
		str = ""
	}

	if err = p.Must(quote); err != nil {
		return nil, err
	}

	return res, nil
}

// SystemLiteral ::= ('"' [^"]* '"') | ("'" [^']* "'")
func (p *parser) parseSystemLiteral() (string, error) {
	var quote rune
	var err error
	if quote, err = p.parseQuote(); err != nil {
		return "", err
	}

	var lit string
	for !p.Test(quote) {
		lit += string(p.Get())
		p.Step()

		if p.isEnd() {
			return "", p.error(fmt.Errorf("could not find quote %c", quote))
		}
	}
	p.Step()

	return lit, nil
}

// PubidLiteral ::= '"' PubidChar* '"' | "'" (PubidChar - "'")* "'"
func (p *parser) parsePubidLiteral() (string, error) {
	var quote rune
	var err error
	if quote, err = p.parseQuote(); err != nil {
		return "", err
	}

	var lit string
	r := p.Get()
	for isPubidChar(r) {
		if r == '\'' && quote == '\'' {
			break
		}
		lit += string(r)
		p.Step()
		r = p.Get()
	}

	if err = p.Must(quote); err != nil {
		return "", err
	}

	return lit, nil
}

/// - Comments

// Comment ::= '<!--' ((Char - '-') | ('-' (Char - '-')))* '-->'
func (p *parser) parseComment() (Comment, error) {
	defer p.setParsing("Comment")()

	if err := p.Musts(`<!--`); err != nil {
		return "", err
	}

	var str Comment
	for !p.Tests("--") {
		r := p.Get()
		if isChar(r) {
			str += Comment(r)
			p.Step()
		} else {
			return "", p.error(errors.New("unexpected character"))
		}
	}

	if err := p.Musts(`-->`); err != nil {
		return "", err
	}

	return str, nil
}

/// - Processing Instructions

// PI ::= ::= '<?' PITarget (S (Char* - (Char* '?>' Char*)))? '?>'
func (p *parser) parsePI() (*PI, error) {
	defer p.setParsing("NOTATION")()

	var err error
	if err = p.Musts("<?"); err != nil {
		return nil, err
	}
	var pi PI
	if pi.Target, err = p.parsePITarget(); err != nil {
		return nil, err
	}

	if isSpace(p.Get()) {
		p.skipSpace()

		for !p.Tests("?>") && !p.isEnd() && isChar(p.Get()) {
			pi.Instruction += string(p.Get())
			p.Step()
		}
	}

	if err = p.Musts("?>"); err != nil {
		return nil, err
	}
	return &pi, nil
}

// PITarget ::= Name - (('X' | 'x') ('M' | 'm') ('L' | 'l'))
func (p *parser) parsePITarget() (string, error) {
	var n string
	var err error
	if n, err = p.parseName(); err != nil {
		return "", err
	}
	if strings.ContainsAny(n, "xmlXML") {
		return "", p.error(errors.New("PI target can not contain 'xml'"))
	}
	return n, nil
}

/// - CDATA Sections

// CDSect ::= CDStart CData CDEnd
func (p *parser) parseCDSect() (CData, error) {
	var err error
	if err = p.Musts("<![CDATA["); err != nil {
		return "", err
	}
	var str string
	for !p.Tests("]]>") {
		if p.isEnd() {
			return "", p.error(errors.New("not found CDSect close tag"))
		}
		str += string(p.Get())
		p.Step()
	}
	p.StepN(len("]]>"))
	return CData(str), nil
}

/// - Document Type Definition

// doctypedecl ::= '<!DOCTYPE' S Name (S ExternalID)? S? ('[' (markupdecl | PEReference | S)* ']' S?)? '>'
func (p *parser) parseDoctype() (*DOCType, error) {
	defer p.setParsing("DOCTYPE")()

	var err error
	if err = p.Musts(`<!DOCTYPE`); err != nil {
		return nil, err
	}
	if err = p.parseSpace(); err != nil {
		return nil, err
	}
	var d DOCType
	d.Name, err = p.parseName()
	if err != nil {
		return nil, err
	}

	p.skipSpace()

	if p.Tests("SYSTEM") || p.Tests("PUBLIC") {
		var ext *ExternalID
		ext, err = p.parseExternalID()
		if err != nil {
			return nil, err
		}
		d.ExtID = ext

		p.skipSpace()
	}

	if p.Test('[') {
		p.Step()

		for {
			if p.Tests("<!ELEMENT") || p.Tests("<!ATTLIST") || p.Tests("<!ENTITY") || p.Tests("<!NOTATION") || p.Tests("<?") || p.Tests("<!--") {
				var m Markup
				m, err = p.parseMarkup()
				if err != nil {
					return nil, err
				}
				d.Markups = append(d.Markups, m)
			} else if p.Test('%') {
				var ref *PERef
				ref, err = p.parsePERef()
				if err != nil {
					return nil, err
				}
				d.PERef = ref
			} else if isSpace(p.Get()) {
				p.skipSpace()
			} else {
				break
			}
		}
		err = p.Must(']')
		if err != nil {
			return nil, err
		}
	}

	p.skipSpace()

	err = p.Must('>')
	if err != nil {
		return nil, err
	}

	return &d, nil
}

// markupdecl ::= elementdecl | AttlistDecl | EntityDecl | NotationDecl | PI | Comment
func (p *parser) parseMarkup() (Markup, error) {
	defer p.setParsing("markup")()

	var err error
	var m Markup
	switch {
	case p.Tests("<!ELEMENT"):
		m, err = p.parseElementDecl()
	case p.Tests("<!ATTLIST"):
		m, err = p.parseAttlist()
	case p.Tests("<!ENTITY"):
		m, err = p.parseEntity()
	case p.Tests("<!NOTATION"):
		m, err = p.parseNotation()
	case p.Tests("<?"):
		m, err = p.parsePI()
	case p.Tests("<!--"):
		m, err = p.parseComment()
	default:
		err = p.error(errors.New("unknown markup"))
	}
	return m, err
}

/// - Standalone Document Declaration

// SDDecl ::= S 'standalone' Eq (("'" ('yes' | 'no') "'") | ('"' ('yes' | 'no') '"'))
func (p *parser) parseStandalone() (bool, error) {
	var err error
	if err = p.parseSpace(); err != nil {
		return false, err
	}
	if err = p.Musts("standalone"); err != nil {
		return false, err
	}
	if err = p.parseEq(); err != nil {
		return false, err
	}
	var quote rune
	if quote, err = p.parseQuote(); err != nil {
		return false, err
	}
	var std bool
	if p.Tests("yes") {
		std = true
		p.StepN(3)
	} else if p.Tests("no") {
		p.StepN(2)
	} else {
		return false, p.error(errors.New("expected 'yes' or 'no'"))
	}
	if err = p.Must(quote); err != nil {
		return false, err
	}
	return std, nil
}

/// - Element

// element ::= EmptyElemTag | STag content ETag
// EmptyElemTag ::= '<' Name (S Attribute)* S? '/>'
// STag ::= '<' Name (S Attribute)* S? '>'
func (p *parser) parseElement() (*Element, error) {
	defer p.setParsing("Element")()

	var err error
	if err = p.Must('<'); err != nil {
		return nil, err
	}
	var e Element

	if e.Name, err = p.parseName(); err != nil {
		return nil, err
	}
	defer p.setParsing(e.Name + " tag")()

	for isSpace(p.Get()) {
		p.skipSpace()
		if p.Test('>') || p.Tests("/>") || p.isEnd() {
			break
		}

		var attr *Attribute
		if attr, err = p.parseAttribute(); err != nil {
			return nil, err
		}
		e.Attrs = append(e.Attrs, attr)
	}

	if p.Test('>') {
		p.Step()

		e.Contents = p.parseContents()

		var endName string
		if endName, err = p.parseETag(); err != nil {
			return nil, err
		}
		if endName != e.Name {
			return nil, p.error(fmt.Errorf("EndTag name %q does not match with StartTag name %q", endName, e.Name))
		}
	} else if p.Tests("/>") {
		p.StepN(len("/>"))
		e.IsEmptyTag = true
	} else {
		return nil, p.error(errors.New("not found element close tag"))
	}
	return &e, nil
}

// Attribute ::= Name Eq AttValue
func (p *parser) parseAttribute() (*Attribute, error) {
	var attr Attribute
	var err error
	if attr.Name, err = p.parseName(); err != nil {
		return nil, err
	}
	if err = p.parseEq(); err != nil {
		return nil, err
	}
	if attr.AttValue, err = p.parseAttValue(); err != nil {
		return nil, err
	}
	return &attr, nil
}

/// - End-tag

// ETag ::= '</' Name S? '>'
func (p *parser) parseETag() (string, error) {
	var n string
	var err error
	if err = p.Musts("</"); err != nil {
		return "", err
	}
	if n, err = p.parseName(); err != nil {
		return "", err
	}
	p.skipSpace()
	if err = p.Must('>'); err != nil {
		return "", err
	}
	return n, nil
}

/// - Content of Elements

func isOnlySpaces(str string) bool {
	for _, r := range []rune(str) {
		if !isSpace(r) {
			return false
		}
	}
	return true
}

// content ::= (element | CharData | Reference | CDSect | PI | Comment)*
func (p *parser) parseContents() []interface{} {
	// '<'Name 			-> Element
	// '&'Name or '&#'	-> Ref
	// '<![CDATA['		-> CDSect
	// '<?'				-> PI
	// '<!--'			-> Comment
	// Not starts with '&' or '<'
	// and not contains ']]>'	-> CharData
	// Others 					-> return

	var res []interface{}
	var err error

	var charData string
	var i interface{}

	for {
		cur := p.cursor

		if p.Test('&') {
			// Ref or break
			i, err = p.parseReference()
			if err != nil {
				p.cursor = cur
				break
			}
			if len(charData) > 0 {
				if !isOnlySpaces(charData) {
					res = append(res, charData)
				}
				charData = ""
			}
			res = append(res, i)
		} else if p.Test('<') {
			if p.Tests("<!") {
				// CDSect or Comment or break
				i, err = p.parseCDSect()
				if err != nil {
					p.cursor = cur

					i, err = p.parseComment()
					if err != nil {
						p.cursor = cur
						break
					}
				}
			} else if p.Tests("<?") {
				// PI or break
				i, err = p.parsePI()
				if err != nil {
					p.cursor = cur
					break
				}
			} else {
				// Element or break
				i, err = p.parseElement()
				if err != nil {
					p.cursor = cur
					break
				}
			}
			if len(charData) > 0 {
				if !isOnlySpaces(charData) {
					res = append(res, charData)
				}
				charData = ""
			}
			res = append(res, i)
		} else {
			if p.isEnd() || p.Tests("]]>") {
				break
			}

			// CharData
			charData += string(p.Get())
			p.Step()
		}
	}
	if len(charData) > 0 {
		if !isOnlySpaces(charData) {
			res = append(res, charData)
		}
	}
	return res
}

/// - Element Type Declaration

// elementdecl ::= '<!ELEMENT' S Name S contentspec S? '>'
func (p *parser) parseElementDecl() (*ElementDecl, error) {
	defer p.setParsing("Element Declaration")()

	var err error
	if err = p.Musts("<!ELEMENT"); err != nil {
		return nil, err
	}
	if err = p.parseSpace(); err != nil {
		return nil, err
	}
	var n string
	if n, err = p.parseName(); err != nil {
		return nil, err
	}
	if err = p.parseSpace(); err != nil {
		return nil, err
	}
	var c ContentSpec
	if c, err = p.parseContentSpec(); err != nil {
		return nil, err
	}
	p.skipSpace()
	if err = p.Must('>'); err != nil {
		return nil, err
	}
	return &ElementDecl{
		Name:        n,
		ContentSpec: c,
	}, nil
}

// contentspec ::= 'EMPTY' | 'ANY' | Mixed | children
func (p *parser) parseContentSpec() (ContentSpec, error) {
	if p.Tests("EMPTY") {
		p.StepN(len("EMPTY"))
		return &EMPTY{}, nil
	} else if p.Tests("ANY") {
		p.StepN(len("ANY"))
		return &ANY{}, nil
	} else {
		var err error

		cur := p.cursor
		{ // try parsing mixed
			var m *Mixed
			m, err = p.parseMixed()
			if err == nil {
				return m, nil
			}
		}
		// reset cursor if it wasn't mixed
		p.cursor = cur

		var ch *Children
		ch, err = p.parseChildren()
		if err != nil {
			return nil, err
		}
		return ch, nil
	}
}

/// - Element-content Models

// children ::= (choice | seq) ('?' | '*' | '+')?
func (p *parser) parseChildren() (*Children, error) {
	var c Children
	var err error

	cur := p.cursor
	{
		var choice *Choice
		choice, err = p.parseChoice()
		if err == nil {
			c.ChoiceSeq = choice
			if p.Test('?') || p.Test('*') || p.Test('+') {
				r := p.Get()
				c.Suffix = &r
				p.Step()
			}
			return &c, nil
		}
	}
	p.cursor = cur

	var s *Seq
	s, err = p.parseSeq()
	if err == nil {
		c.ChoiceSeq = s
		if p.Test('?') || p.Test('*') || p.Test('+') {
			r := p.Get()
			c.Suffix = &r
			p.Step()
		}
		return &c, nil
	}

	return nil, p.error(errors.New("unknown error"))
}

// cp ::= (Name | choice | seq) ('?' | '*' | '+')?
func (p *parser) parseCP() (*CP, error) {
	var cp CP
	var err error
	if p.Test('(') { // choice or seq
		cur := p.cursor

		var choice *Choice
		choice, err = p.parseChoice()
		if err != nil {
			p.cursor = cur

			var seq *Seq
			seq, err = p.parseSeq()
			if err != nil {
				return nil, err
			}
			cp.ChoiceSeq = seq
		} else {
			cp.ChoiceSeq = choice
		}
	} else {
		var n string
		n, err = p.parseName()
		if err != nil {
			return nil, err
		}
		cp.Name = n
	}

	if p.Test('?') || p.Test('*') || p.Test('+') {
		r := p.Get()
		cp.Suffix = &r
		p.Step()
	}

	return &cp, nil
}

// choice ::= '(' S? cp ( S? '|' S? cp )* S? ')'
func (p *parser) parseChoice() (*Choice, error) {
	if err := p.Must('('); err != nil {
		return nil, err
	}
	p.skipSpace()
	var cps []CP
	cp, err := p.parseCP()
	if err != nil {
		return nil, err
	}
	cps = append(cps, *cp)
	for {
		p.skipSpace()
		if !p.Test('|') {
			break
		}
		p.Step()

		cp, err = p.parseCP()
		if err != nil {
			return nil, err
		}
		cps = append(cps, *cp)
	}
	if err := p.Must(')'); err != nil {
		return nil, err
	}
	return &Choice{
		CPs: cps,
	}, nil
}

// seq ::= '(' S? cp ( S? ',' S? cp )* S? ')'
func (p *parser) parseSeq() (*Seq, error) {
	if err := p.Must('('); err != nil {
		return nil, err
	}
	p.skipSpace()
	var cps []CP
	cp, err := p.parseCP()
	if err != nil {
		return nil, err
	}

	cps = append(cps, *cp)
	for {
		p.skipSpace()
		if !p.Test(',') {
			break
		}
		p.Step()

		cp, err = p.parseCP()
		if err != nil {
			return nil, err
		}
		cps = append(cps, *cp)
	}

	if err := p.Must(')'); err != nil {
		return nil, err
	}
	return &Seq{
		CPs: cps,
	}, nil
}

/// - Mixed-content Declaration

// Mixed ::= '(' S? '#PCDATA' (S? '|' S? Name)* S? ')*' | '(' S? '#PCDATA' S? ')'
func (p *parser) parseMixed() (*Mixed, error) {
	if err := p.Must('('); err != nil {
		return nil, err
	}
	p.skipSpace()
	if err := p.Musts("#PCDATA"); err != nil {
		return nil, err
	}

	var m Mixed
	for {
		p.skipSpace()
		if p.Test(')') {
			p.Step()
			break
		}
		if p.isEnd() {
			return nil, p.error(errors.New(`could not find ')'`))
		}
		var err error
		if err = p.Must('|'); err != nil {
			return nil, err
		}
		p.skipSpace()
		var n string
		n, err = p.parseName()
		if err != nil {
			return nil, err
		}
		m.Names = append(m.Names, n)
	}

	if len(m.Names) > 0 {
		if err := p.Must('*'); err != nil {
			return nil, err
		}
	}

	return &m, nil
}

/// - Attribute-list Declaration

// AttlistDecl ::= '<!ATTLIST' S Name AttDef* S? '>'
func (p *parser) parseAttlist() (*Attlist, error) {
	defer p.setParsing("ATTLIST")()

	var att Attlist
	var err error
	if err = p.Musts("<!ATTLIST"); err != nil {
		return nil, err
	}
	if err = p.parseSpace(); err != nil {
		return nil, err
	}
	if att.Name, err = p.parseName(); err != nil {
		return nil, err
	}
	// S Name  or  S? '>'
	for {
		cur := p.cursor

		p.skipSpace()
		if p.Test('>') {
			p.Step()
			break
		}
		if p.isEnd() {
			return nil, p.error(errors.New("reached EOF"))
		}
		p.cursor = cur

		var def *AttDef
		if def, err = p.parseAttDef(); err != nil {
			return nil, err
		}
		att.Defs = append(att.Defs, def)
	}

	return &att, nil
}

// AttDef ::= S Name S AttType S DefaultDecl
func (p *parser) parseAttDef() (*AttDef, error) {
	var err error
	if err = p.parseSpace(); err != nil {
		return nil, err
	}

	var def AttDef
	if def.Name, err = p.parseName(); err != nil {
		return nil, err
	}

	if err = p.parseSpace(); err != nil {
		return nil, err
	}

	if def.Type, err = p.parseAttType(); err != nil {
		return nil, err
	}

	if err = p.parseSpace(); err != nil {
		return nil, err
	}

	if def.Decl, err = p.parseDefaultDecl(); err != nil {
		return nil, err
	}

	return &def, nil
}

/// - Attribute Types

// AttType ::= StringType | TokenizedType | EnumeratedType
func (p *parser) parseAttType() (AttType, error) {
	if p.Tests(AttTokenCDATA.ToString()) {
		p.StepN(len(AttTokenCDATA.ToString()))
		return AttTokenCDATA, nil
	} else if p.Tests(AttTokenID.ToString()) || p.Tests(AttTokenIDREF.ToString()) || p.Tests(AttTokenIDREFS.ToString()) || p.Tests(AttTokenENTITY.ToString()) || p.Tests(AttTokenENTITIES.ToString()) || p.Tests(AttTokenNMTOKEN.ToString()) || p.Tests(AttTokenNMTOKENS.ToString()) {
		var tok AttToken
		switch {
		case p.Tests(AttTokenID.ToString()):
			tok = AttTokenID
			if p.Tests(AttTokenIDREF.ToString()) {
				tok = AttTokenIDREF
				if p.Tests(AttTokenIDREFS.ToString()) {
					tok = AttTokenIDREFS
				}
			}
		case p.Tests(AttTokenENTITY.ToString()):
			tok = AttTokenENTITY
		case p.Tests(AttTokenENTITIES.ToString()):
			tok = AttTokenENTITIES
		case p.Tests(AttTokenNMTOKEN.ToString()):
			tok = AttTokenNMTOKEN
			if p.Tests(AttTokenNMTOKENS.ToString()) {
				tok = AttTokenNMTOKENS
			}
		}
		p.StepN(len(tok.ToString()))
		return tok, nil
	} else if p.Tests("NOTATION") {
		return p.parseNotationType()
	} else if p.Test('(') {
		return p.parseEnum()
	}
	return nil, p.error(errors.New("unknwon error"))
}

/// - Enumerated Attribute Types

// NotationType ::= 'NOTATION' S '(' S? Name (S? '|' S? Name)* S? ')'
func (p *parser) parseNotationType() (*NotationType, error) {
	var err error
	if err = p.Musts("NOTATION"); err != nil {
		return nil, err
	}
	if err = p.parseSpace(); err != nil {
		return nil, err
	}
	if err = p.Must('('); err != nil {
		return nil, err
	}
	p.skipSpace()

	var n NotationType
	var t string
	if t, err = p.parseName(); err != nil {
		return nil, err
	}
	n.Names = append(n.Names, t)

	for {
		cur := p.cursor

		p.skipSpace()
		if p.Test(')') {
			p.Step()
			break
		}
		if p.isEnd() {
			return nil, p.error(errors.New(`expected ')'`))
		}
		p.cursor = cur

		p.skipSpace()
		if err = p.Must('|'); err != nil {
			return nil, err
		}
		p.skipSpace()

		if t, err = p.parseName(); err != nil {
			return nil, err
		}
		n.Names = append(n.Names, t)
	}

	return &n, nil
}

// Enumeration ::= '(' S? Nmtoken (S? '|' S? Nmtoken)* S? ')'
func (p *parser) parseEnum() (*Enum, error) {
	var err error
	var e Enum

	if err = p.Must('('); err != nil {
		return nil, err
	}
	p.skipSpace()

	var nm string
	if nm, err = p.parseNmtoken(); err != nil {
		return nil, err
	}
	e.Cases = append(e.Cases, nm)

	for {
		cur := p.cursor

		p.skipSpace()
		if p.Test(')') {
			p.Step()
			break
		}
		if p.isEnd() {
			return nil, p.error(errors.New(`expected ')'`))
		}
		p.cursor = cur

		p.skipSpace()
		if err = p.Must('|'); err != nil {
			return nil, err
		}
		p.skipSpace()

		if nm, err = p.parseNmtoken(); err != nil {
			return nil, err
		}
		e.Cases = append(e.Cases, nm)
	}

	return &e, nil
}

// Nmtoken ::= (NameChar)+
func (p *parser) parseNmtoken() (string, error) {
	var str string
	r := p.Get()
	for isNameChar(r) {
		str += string(r)
		p.Step()
		r = p.Get()
	}
	if len(str) == 0 {
		return "", p.error(errors.New("empty Nmtoken"))
	}
	return str, nil
}

/// - Attribute Defaults

// DefaultDecl ::= '#REQUIRED' | '#IMPLIED' | (('#FIXED' S)? AttValue)
func (p *parser) parseDefaultDecl() (*DefaultDecl, error) {
	var d DefaultDecl
	var err error
	if p.Tests(DefaultDeclTypeRequired.ToString()) {
		p.StepN(len(DefaultDeclTypeRequired.ToString()))
		d.Type = DefaultDeclTypeRequired
		return &d, nil
	} else if p.Tests(DefaultDeclTypeImplied.ToString()) {
		p.StepN(len(DefaultDeclTypeImplied.ToString()))
		d.Type = DefaultDeclTypeImplied
		return &d, nil
	} else {
		if p.Tests(DefaultDeclTypeFixed.ToString()) {
			p.StepN(len(DefaultDeclTypeFixed.ToString()))
			if err = p.parseSpace(); err != nil {
				return nil, err
			}
		}
		d.Type = DefaultDeclTypeFixed
		if d.AttValue, err = p.parseAttValue(); err != nil {
			return nil, err
		}
		return &d, nil
	}
}

/// - Character Reference

// CharRef ::= '&#' [0-9]+ ';' | '&#x' [0-9a-fA-F]+ ';'
func (p *parser) parseCharRef() (*CharRef, error) {
	var ref CharRef
	var err error

	if p.Tests("&#x") {
		ref.Prefix = "&#x"
		p.StepN(len("&#x"))

		r := p.Get()
		if !isNum(r) && !isAlpha(r) {
			return nil, p.error(errors.New("expected number or alphabet character"))
		}

		for isNum(r) || isAlpha(r) {
			ref.Value += string(r)
			p.Step()
			r = p.Get()
		}
	} else if p.Tests("&#") {
		ref.Prefix = "&#"
		p.StepN(len("&#"))

		r := p.Get()
		if !isNum(r) {
			return nil, p.error(errors.New("expected number"))
		}

		for isNum(r) {
			ref.Value += string(r)
			p.Step()
			r = p.Get()
		}
	} else {
		return nil, p.error(errors.New("expected '&#x' or '&#'"))
	}

	if err = p.Must(';'); err != nil {
		return nil, err
	}

	return &ref, nil
}

/// - Entity Reference

// Reference ::= EntityRef | CharRef
func (p *parser) parseReference() (Ref, error) {
	var err error
	cur := p.cursor

	// try EntityRef
	var eRef *EntityRef
	eRef, err = p.parseEntityRef()
	if err != nil {
		// reset
		p.cursor = cur
		err = nil

		// try CharRef
		var cRef *CharRef
		cRef, err = p.parseCharRef()
		if err != nil {
			return nil, err
		}
		return cRef, nil
	}
	return eRef, nil
}

// EntityRef ::= '&' Name ';'
func (p *parser) parseEntityRef() (*EntityRef, error) {
	var err error
	if err = p.Must('&'); err != nil {
		return nil, err
	}
	var e EntityRef
	e.Name, err = p.parseName()
	if err != nil {
		return nil, err
	}
	if err = p.Must(';'); err != nil {
		return nil, err
	}
	return &e, nil
}

// PEReference ::= '%' Name ';'
func (p *parser) parsePERef() (*PERef, error) {
	var err error
	if err = p.Must('%'); err != nil {
		return nil, err
	}
	var e PERef
	e.Name, err = p.parseName()
	if err != nil {
		return nil, err
	}
	if err = p.Must(';'); err != nil {
		return nil, err
	}
	return &e, nil
}

/// - Entity Declaration

// EntityDecl ::= GEDecl |  PEDecl
func (p *parser) parseEntity() (*Entity, error) {
	defer p.setParsing("Entity Declaration")()

	var err error
	if err = p.Musts("<!ENTITY"); err != nil {
		return nil, err
	}
	if err = p.parseSpace(); err != nil {
		return nil, err
	}

	var e Entity

	// PEDecl ::= '<!ENTITY' S '%' S Name S PEDef S? '>'
	// GEDecl ::= '<!ENTITY' S Name S EntityDef S? '>'

	if p.Test('%') {
		e.Type = EntityTypePE

		p.Step()

		if err = p.parseSpace(); err != nil {
			return nil, err
		}
	} else {
		e.Type = EntityTypeGE
	}

	if e.Name, err = p.parseName(); err != nil {
		return nil, err
	}

	if err = p.parseSpace(); err != nil {
		return nil, err
	}

	if e.Type == EntityTypePE {
		// PEDef
		e.Value, e.ExtID, err = p.parsePEDef()
	} else {
		// EntityDef
		e.Value, e.ExtID, e.NData, err = p.parseEntityDef()
	}
	if err != nil {
		return nil, err
	}

	p.skipSpace()

	if err = p.Must('>'); err != nil {
		return nil, err
	}

	return &e, nil
}

// EntityDef ::= EntityValue | (ExternalID NDataDecl?)
func (p *parser) parseEntityDef() (EntityValue, *ExternalID, string, error) {
	var value EntityValue
	var ndata string
	var ext *ExternalID
	var err error

	if p.Test('\'') || p.Test('"') {
		value, err = p.parseEntityValue()
		if err != nil {
			return nil, nil, "", err
		}
		return value, nil, "", nil
	} else if p.Tests("SYSTEM") || p.Tests("PUBLIC") {
		ext, err = p.parseExternalID()
		if err != nil {
			return nil, nil, "", err
		}

		cur := p.cursor

		ndata, err = p.parseNData()
		if err != nil {
			p.cursor = cur
		}

		return nil, ext, ndata, nil
	} else {
		return nil, nil, "", p.error(errors.New("unexpected char"))
	}
}

// PEDef ::= EntityValue | ExternalID
func (p *parser) parsePEDef() (EntityValue, *ExternalID, error) {
	var value EntityValue
	var ext *ExternalID
	var err error

	if p.Test('\'') || p.Test('"') {
		value, err = p.parseEntityValue()
		if err != nil {
			return nil, nil, err
		}
		return value, nil, nil
	} else if p.Tests("SYSTEM") || p.Tests("PUBLIC") {
		ext, err = p.parseExternalID()
		if err != nil {
			return nil, nil, err
		}

		return nil, ext, nil
	} else {
		return nil, nil, p.error(errors.New("unexpected char"))
	}
}

/// - External Entity Declaration

// ExternalID ::= 'SYSTEM' S SystemLiteral | 'PUBLIC' S PubidLiteral S SystemLiteral
func (p *parser) parseExternalID() (*ExternalID, error) {
	var ext ExternalID
	if p.Tests("SYSTEM") {
		p.StepN(len("SYSTEM"))
		ext.Type = ExternalTypeSystem
	} else if p.Tests("PUBLIC") {
		p.StepN(len("PUBLIC"))
		ext.Type = ExternalTypePublic
	} else {
		return nil, p.error(errors.New("expected 'SYSTEM' or 'PUBLIC'"))
	}

	if err := p.parseSpace(); err != nil {
		return nil, err
	}

	if ext.Type == ExternalTypePublic {
		pubid, err := p.parsePubidLiteral()
		if err != nil {
			return nil, err
		}
		ext.Pubid = pubid

		if err := p.parseSpace(); err != nil {
			return nil, err
		}
	}

	sys, err := p.parseSystemLiteral()
	if err != nil {
		return nil, err
	}
	ext.System = sys

	return &ext, nil
}

// NDataDecl ::= S 'NDATA' S Name
func (p *parser) parseNData() (string, error) {
	var err error
	if err = p.parseSpace(); err != nil {
		return "", err
	}
	if err = p.Musts("NDATA"); err != nil {
		return "", err
	}
	if err = p.parseSpace(); err != nil {
		return "", err
	}
	return p.parseName()
}

/// - Encoding Declaration

// EncodingDecl ::= S 'encoding' Eq ('"' EncName  '"' |  "'" EncName "'" )
func (p *parser) parseEncoding() (string, error) {
	var err error
	if err = p.parseSpace(); err != nil {
		return "", err
	}
	if err = p.Musts("encoding"); err != nil {
		return "", err
	}
	if err = p.parseEq(); err != nil {
		return "", err
	}

	var quote rune
	if quote, err = p.parseQuote(); err != nil {
		return "", err
	}

	var enc string
	enc, err = p.parseEncName()
	if err != nil {
		return "", err
	}

	if err = p.Must(quote); err != nil {
		return "", err
	}

	return enc, nil
}

// EncName ::= [A-Za-z] ([A-Za-z0-9._] | '-')*
func (p *parser) parseEncName() (string, error) {
	var str string
	r := p.Get()
	if !isAlpha(r) {
		return "", p.error(errors.New("Encoding name contains non alphabet"))
	}
	str += string(r)
	p.Step()

	for {
		if isAlpha(p.Get()) || isNum(p.Get()) || p.Test('.') || p.Test('_') || p.Test('-') {
			str += string(p.Get())
			p.Step()
		} else {
			break
		}
	}

	return str, nil
}

/// - Notation Declarations

// NotationDecl ::= '<!NOTATION' S Name S (ExternalID | PublicID) S? '>'
func (p *parser) parseNotation() (*Notation, error) {
	defer p.setParsing("NOTATION")()

	var n Notation
	var err error
	if err = p.Musts("<!NOTATION"); err != nil {
		return nil, err
	}
	if err = p.parseSpace(); err != nil {
		return nil, err
	}
	if n.Name, err = p.parseName(); err != nil {
		return nil, err
	}
	if err = p.parseSpace(); err != nil {
		return nil, err
	}

	// ExternalID ::= 'SYSTEM' S SystemLiteral | 'PUBLIC' S PubidLiteral
	var ext ExternalID
	if p.Tests("SYSTEM") {
		p.StepN(len("SYSTEM"))
		ext.Type = ExternalTypeSystem
	} else if p.Tests("PUBLIC") {
		p.StepN(len("PUBLIC"))
		ext.Type = ExternalTypePublic
	} else {
		return nil, p.error(errors.New("expected 'SYSTEM' or 'PUBLIC'"))
	}

	if err := p.parseSpace(); err != nil {
		return nil, err
	}

	if ext.Type == ExternalTypePublic {
		pubid, err := p.parsePubidLiteral()
		if err != nil {
			return nil, err
		}
		ext.Pubid = pubid

		// ( S SystemLiteral )?
		cur := p.cursor

		err = p.parseSpace()
		if err == nil {
			var sys string
			sys, err = p.parseSystemLiteral()
			if err == nil {
				ext.System = sys
			} else {
				err = nil
				p.cursor = cur
			}
		} else {
			p.cursor = cur
			err = nil
		}
	} else {
		var sys string
		sys, err = p.parseSystemLiteral()
		if err != nil {
			return nil, err
		}
		ext.System = sys
	}

	n.ExtID = ext

	p.skipSpace()

	if err = p.Must('>'); err != nil {
		return nil, err
	}

	return &n, nil
}

/// - Others

func (p *parser) parseQuote() (rune, error) {
	var err error
	r := p.Get()
	if isQuote(r) {
		p.Step()
	} else {
		err = p.error(errors.New(`expected ' or "`))
	}
	return r, err
}
