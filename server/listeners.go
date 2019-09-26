package server

type listeners struct {
	onDoneCrawling []func(int, int)
	onDoneLoading  []func(int, int)
	onDoneIndexing []func(int, int)
	onError        []func(error)
}

func (this *listeners) OnDoneCrawling(listener func(int, int)) {
	this.onDoneCrawling = append(this.onDoneCrawling, listener)
}

func (this *listeners) OnDoneLoading(listener func(int, int)) {
	this.onDoneLoading = append(this.onDoneLoading, listener)
}

func (this *listeners) OnDoneIndexing(listener func(int, int)) {
	this.onDoneIndexing = append(this.onDoneIndexing, listener)
}

func (this *listeners) OnError(listener func(error)) {
	this.onError = append(this.onError, listener)
}

func (this *listeners) DoneCrawling(fileCount, byteCount int) {
	for _, listener := range this.onDoneCrawling {
		go listener(fileCount, byteCount)
	}
}

func (this *listeners) DoneLoading(fileCount, byteCount int) {
	for _, listener := range this.onDoneLoading {
		go listener(fileCount, byteCount)
	}
}

func (this *listeners) DoneIndexing(fileCount, byteCount int) {
	for _, listener := range this.onDoneIndexing {
		go listener(fileCount, byteCount)
	}
}

func (this *listeners) Error(err error) {
	for _, listener := range this.onError {
		go listener(err)
	}
}
