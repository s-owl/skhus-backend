package main

import (
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
)

// GetProcessInfo 프로세스 정보를 가져온다.
func GetProcessInfo(c *gin.Context) {
	out, err := exec.Command("ps", "-eo", "user,pid,ppid,rss,size,vsize,pmem,pcpu,time,cmd").CombinedOutput()
	if err != nil {
		c.String(http.StatusInternalServerError, "명령어 실행 실패")
		return
	}
	c.Data(http.StatusOK, "text/plain", out)
}
