package pdfmix

const (
	//PermissionsPrint setProtection print
	PermissionsPrint = 4
	//PermissionsModify setProtection modify
	PermissionsModify = 8
	//PermissionsCopy setProtection copy
	PermissionsCopy = 16
	//PermissionsAnnotForms setProtection  annot-forms
	PermissionsAnnotForms = 32
	//PermissionsAll all
	PermissionsAll = PermissionsPrint | PermissionsModify | PermissionsCopy | PermissionsAnnotForms
)

//PasswordInfo protect pdf with password
type PasswordInfo struct {
	Permissions int
	OwnerPass   []byte
	UserPass    []byte
}
