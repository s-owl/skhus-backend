package consts

import (
	"os"
	"strings"
)

const ForestURL string = "https://forest.skhu.ac.kr"
const ForestDomain string = "forest.skhu.ac.kr"
const SkhuSamURL string = "http://sam.skhu.ac.kr"
const SkhuCasURL string = "http://cas.skhu.ac.kr"
const SkhuURL string = "http://skhu.ac.kr"

const CoreSecurity string = "*/CoreSecurity.js"

const UserAgentIE string = "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/6.0)"
const UserAgentMacOsChrome string = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36"

const CredentialMalformedMsg string = "Empty or malformed credential data.\n비어 있거나 올바르지 않은 인증 데이터 입니다."
const InternalError string = "An internal error occured while processing data\n데이터 처리중 내부적 오류가 발생했습니다."

func IsDebug() bool {
	debug := strings.ToLower(os.Getenv("DEBUG"))
	return strings.Compare(debug, "true") == 0
}

func SkhusWebSite() []string {
	if IsDebug() {
		return []string{"http://localhost:3000"}
	}
	return []string{"https://skhus-web.sleepy-owl.com"}
}
