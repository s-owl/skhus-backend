# skhus-backend

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/41514677725048c58b8a2ec79671e936)](https://app.codacy.com/app/s-owl/skhus-backend?utm_source=github.com&utm_medium=referral&utm_content=s-owl/skhus-backend&utm_campaign=Badge_Grade_Settings)

SKHU's 서비스의 API 백엔드 입니다. [Go][1], [Gin][2], [GoQuery][3], [Chromedp][4] 로 개발 되었으며, [기존 Node.js 기반의 백엔드][6]를 대체합니다.  
[Chromdep][4]의 [headless-shell][5] 과, [Go][1]를 활용하여 개발한 덕에, [기존 Node.js 기반 백엔드][6]에 비해 차지하는 용량과 리소스 사용량은 매우 낮으면서도, 높은 성능을 발휘 합니다.

The API Backend for SKHU's Service built with [Go][1], [Gin][2], [GoQuery][3] and [Chromedp][4]. And It replaces [legacy Node.js based backend][6].
Thanks to [Chromdep][4]'s [headless-shell][5] and [Go][1]. It has much higher performance with very low resource and storage usage than [old Node.js based backend][6]

## 바로 실행하기
Chrome 이 먼저 시스템에 설치 되어 있어야 합니다.  
Chrome is required to be installed on your system.
```bash
git clone https://github.com/s-owl/skhus-backend.git
cd skhus-backend
go run .
```

## Docker 로 실행
```bash
docker build . --file Dockerfile --tag skhus-backend:latest
docker run skhus-backend:latest
```
[미리 빌드된 이미지][7]를 사용하여 실행도 가능합니다.   
You can also run with [prebuilt images][7].   
```bash
docker pull docker.pkg.github.com/s-owl/skhus-backend/backend:[tag]
docker run docker.pkg.github.com/s-owl/skhus-backend/backend:[tag]
```
[1]: https://golang.org
[2]: https://github.com/gin-gonic/gin
[3]: https://github.com/PuerkitoBio/goquery
[4]: https://github.com/chromedp/chromedp
[5]: https://github.com/chromedp/docker-headless-shell
[6]: https://github.com/s-owl/skhu-backend
[7]: https://github.com/s-owl/skhus-backend/packages/21651
