package mail

import "github.com/chiefsend/api/models"

type Mail interface {
	SendMail(share models.Share)
}
