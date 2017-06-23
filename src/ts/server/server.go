package main

var BetaTokenList = []string{
	"b786b9ca7098a4174c3d1cd6fe96b6bc",
	"664c9f6d06599bb530a254247a3b5e0b",
	"90de8e0fb09ad0deeb05c4a936113f5e",
	"2c9cc67d4d1b3e3f74490299a1daa6d4", //anton
	"6fac90fbc172a584ea245d2d8f2c4b3c", //anton
	"bb76e3afd6cbd79e5d635e1abef44401", //evg
	"8f13167760196046f3bbb2592b1c7e79", //evg
}

func IsDebug(token string) bool {
	for _, t := range BetaTokenList {
		if t == token {
			return true
		}
	}
	return false
}
