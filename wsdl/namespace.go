package wsdl

// NamespacePrefixMap returns a mapping of XML namespace prefixes to URIs
// from the xmlns:* declarations on the <definitions> element.
// For example, xmlns:ts="http://example.com/data" yields {"ts": "http://example.com/data"}.
func (d *Definitions) NamespacePrefixMap() map[string]string {
	m := make(map[string]string)
	for _, attr := range d.ExtraAttrs {
		if attr.Name.Space == "xmlns" {
			m[attr.Name.Local] = attr.Value
		}
	}
	return m
}

// ResolveQName resolves a QName reference (e.g. "ts:addPolicyRequest") using
// the xmlns declarations on <definitions>. Returns the namespace URI and
// local name. If the reference has no prefix, the namespace URI falls back
// to the WSDL's targetNamespace.
func (d *Definitions) ResolveQName(ref string) (nsURI, localName string) {
	for i := len(ref) - 1; i >= 0; i-- {
		if ref[i] == ':' {
			prefix := ref[:i]
			localName = ref[i+1:]
			if uri, ok := d.NamespacePrefixMap()[prefix]; ok {
				return uri, localName
			}
			return "", localName
		}
	}
	return d.TargetNamespace, ref
}
