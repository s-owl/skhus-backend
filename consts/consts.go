package consts

import (
	"os"
	"strings"
)

// 도메인 상수 값 지정

// (구)학사행정시스템의 URL
const ForestURL string = "https://forest.skhu.ac.kr"
// (구)학사행정시스템의 Domain
const ForestDomain string = "forest.skhu.ac.kr"
// (신)학사행정시스템의 URL
const SkhuSamURL string = "http://sam.skhu.ac.kr"
// (신)학사행정시스템의 URL
const SkhuCasURL string = "http://cas.skhu.ac.kr"
// 대학교 홈페이지 주소
const SkhuURL string = "http://skhu.ac.kr"

// 자바스크립트 보안 스크립트
const CoreSecurity string = "*/CoreSecurity.js"

// 임의 브라우저 설정
const UserAgentIE string = "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/6.0)"
// 임의 브라우저 설정
const UserAgentMacOsChrome string = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36"

// 인증 데이터 오류 메세지
const CredentialMalformedMsg string = "Empty or malformed credential data.\n비어 있거나 올바르지 않은 인증 데이터 입니다."
// 내부 오류 메세지
const InternalError string = "An internal error occured while processing data\n데이터 처리중 내부적 오류가 발생했습니다."

// 디버그 모드 확인 함수
func IsDebug() bool {
	debug := strings.ToLower(os.Getenv("DEBUG"))
	return strings.Compare(debug, "true") == 0
}

// 웹 버전의 URL을 리턴하는 함수
func SkhusWebSite() []string {
	if IsDebug() {
		return []string{"http://localhost:3000"}
	}
	return []string{"https://skhus-web.sleepy-owl.com"}
}
