package types

type ConfirmationRes struct {
	Status string
}

type DevicePostRes struct {
	ID            uint
	Name          string
	DateAddedInMs uint
	DeviceToken   string
}

type DeviceRegistrationRes struct {
	ID           uint
	IsRegistered bool
}
