package git

// PullRequestInput represents the input data when creating a
// pull request.
type PullRequestInput struct {
	RepositoryURL string
	Title         string
	Body          string
	// Name of the branch that implements the changes.
	Head string
	// Name of the branch that will receive the changes.
	Base    string
	Commits []Commit
}

// PullRequest represents the result after successfully
// creating a pull request.
type PullRequest struct {
	// Title is the title of the pull request.
	Title string
	// Description is the description of the pull request.
	Description string
	// Link links to the pull request page.
	Link string
	// Merge shows if the pull request is merged or not.
	Merged bool
}
