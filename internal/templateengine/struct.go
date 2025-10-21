package templateengine

// ArchetypeFile holds the relative path of the file from the archetypes folder, and the contents of the file
// with the provided values for the template (if any)
type ArchetypeFile struct {
	Name    string
	Content []byte
}
