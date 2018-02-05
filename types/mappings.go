package types

type MountPoint struct {
	Source string
	Target string
}

type MountPointsByRegistry map[string][]MountPoint
