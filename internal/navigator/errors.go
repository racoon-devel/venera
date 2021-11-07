package navigator

import (
	"path"

	uuid "github.com/satori/go.uuid"
)

func (n *Navigator) SetErrorReportsPath(path string) *Navigator {
	n.dumpPath = path
	return n
}

func (n *Navigator) checkError(method string) {
	if n.err == nil {
		return
	}

	tmpUUID := uuid.NewV4().String()

	if n.batch != "" {
		n.log.Errorf("batch '%s' error: %s failed: %+v [ %s ]", n.batch, method, n.err)
	} else {
		n.log.Errorf("browser error: %s failed: %+v [ %s ]", method, n.err, tmpUUID)
	}

	n.Screenshot(path.Join(n.dumpPath, tmpUUID+".jpg")).TracePage(path.Join(n.dumpPath, tmpUUID+".html"))
}
