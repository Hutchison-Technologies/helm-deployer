package runtime

func PanicIfError(e error) {
	if e != nil {
		panic(e.Error())
	}
}
