package gee

func assert1(flag bool, text string) {
	if !flag {
		panic(text)
	}
}
