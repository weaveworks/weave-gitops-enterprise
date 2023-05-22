package git

type jscmRefUpdate struct {
	Name        string `json:"name"`
	OldObjectID string `json:"oldObjectId,omitempty"`
}

type jscmChange struct {
	ChangeType string `json:"changeType"`
	Item       struct {
		Path string `json:"path"`
	} `json:"item"`
	NewContent struct {
		Content     string `json:"content,omitempty"`
		ContentType string `json:"contentType,omitempty"`
	} `json:"newContent,omitempty"`
}

type jscmCommitRef struct {
	Comment  string       `json:"comment,omitempty"`
	Changes  []jscmChange `json:"changes,omitempty"`
	CommitID string       `json:"commitId"`
	URL      string       `json:"url"`
}

type jscmContentCreateUpdate struct {
	RefUpdates []jscmRefUpdate `json:"refUpdates"`
	Commits    []jscmCommitRef `json:"commits"`
}

type jscmContent struct {
	ObjectID      string `json:"objectId"`
	GitObjectType string `json:"gitObjectType"`
	CommitID      string `json:"commitId"`
	Path          string `json:"path"`
	Content       string `json:"content"`
	URL           string `json:"url"`
	Links         struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Repository struct {
			Href string `json:"href"`
		} `json:"repository"`
		Blob struct {
			Href string `json:"href"`
		} `json:"blob"`
	} `json:"_links"`
}

type jscmContentList struct {
	Count int            `json:"count"`
	Value []*jscmContent `json:"value"`
}
