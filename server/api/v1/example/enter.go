package example

import "github.com/flipped-aurora/gin-vue-admin/server/service"

type ApiGroup struct {
	AttachmentCategoryApi
	FileUploadAndDownloadApi
}

var (
	attachmentCategoryService    = service.ServiceGroupApp.ExampleServiceGroup.AttachmentCategoryService
	fileUploadAndDownloadService = service.ServiceGroupApp.ExampleServiceGroup.FileUploadAndDownloadService
)
