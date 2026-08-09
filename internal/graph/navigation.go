package graph

// Navigation holds navigation elements for a diagram.
// It is nil for C1 root diagrams and populated for C2/C3 diagrams.
type Navigation struct {
	// BackLink is the parent back-link (nil for C1 root).
	BackLink *BackLink
	// Breadcrumbs is the breadcrumb trail from root to current.
	Breadcrumbs []BreadcrumbItem
}

// BackLink represents a link back to the parent diagram.
type BackLink struct {
	// Name is the display name for the parent.
	Name string
	// URL is the relative path to the parent diagram.
	URL string
}

// BreadcrumbItem represents one item in the breadcrumb trail.
type BreadcrumbItem struct {
	// Name is the display name.
	Name string
	// URL is the relative path (empty for current level).
	URL string
}
