package struct2csv

import (
	"strings"
)

const (
	separator          = '/'
	builderDefaultSize = 64
)

// PathBuilder is a sequence of Token.
type PathBuilder struct {
	b *strings.Builder
}

func NewPathBuilder(growSize int) PathBuilder {
	b := &strings.Builder{}
	if growSize == 0 {
		growSize = builderDefaultSize
	}
	b.Grow(growSize)
	return PathBuilder{b: b}
}

// AppendString appends the token.
func (p PathBuilder) AppendString(token string) PathBuilder {
	p.b.WriteRune(separator)
	p.b.WriteString(token)
	return p
}

// Clone returns a duplicate of the PathBuilder.
func (p PathBuilder) Clone(growSize int) PathBuilder {
	if p.b.Len() == 0 {
		return NewPathBuilder(growSize)
	}

	b := &strings.Builder{}
	if growSize == 0 {
		growSize = builderDefaultSize
	}
	b.Grow(growSize)
	b.WriteString(p.b.String())
	return PathBuilder{b}
}

// String returns full path string.
func (p PathBuilder) String() string {
	s := p.b.String()
	p.b.Reset()
	return s
}
