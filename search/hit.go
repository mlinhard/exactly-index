package search

type HitStruct struct {
	searchResult SearchResult
	hitIdx       int
}

type HitContextStruct struct {
	data       []byte
	position   int32
	lenBefore  int32
	lenPattern int32
	lenAfter   int32
}

func (this *HitStruct) GlobalPosition() int {
	return this.searchResult.globalPosition(this.hitIdx)
}

func (this *HitStruct) Position() int {
	return this.searchResult.position(this.hitIdx)
}

func (this *HitStruct) Document() *Document {
	return this.searchResult.document(this.hitIdx)
}

func (this *HitStruct) CharContext(charsBefore, charsAfter int) HitContext {
	return this.searchResult.charContext(this.hitIdx, charsAfter, charsAfter)
}

func (this *HitStruct) LineContext(linesBefore, linesAfter int) HitContext {
	return this.searchResult.lineContext(this.hitIdx, linesBefore, linesAfter)
}

func (this *HitContextStruct) Before() []byte {
	return this.data[this.position : this.position+this.lenBefore]
}

func (this *HitContextStruct) Pattern() []byte {
	start := this.position + this.lenBefore
	return this.data[start : start+this.lenPattern]
}

func (this *HitContextStruct) After() []byte {
	start := this.position + this.lenBefore + this.lenPattern
	return this.data[start : start+this.lenAfter]
}

func (this *HitContextStruct) HighlightStart() int {
	return int(this.lenBefore)
}

func (this *HitContextStruct) HighlightEnd() int {
	return int(this.lenBefore + this.lenPattern)
}
