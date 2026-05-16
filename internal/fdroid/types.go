package fdroid

type IndexV1 struct {
	Repo     RepoV1               `json:"repo"`
	Apps     []AppV1              `json:"apps"`
	Packages map[string][]PackageV1 `json:"packages"`
}

type RepoV1 struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	Description string `json:"description"`
	Version     int    `json:"version"`
}

type AppV1 struct {
	PackageName string                        `json:"packageName"`
	Name        string                        `json:"name"`
	Summary     string                        `json:"summary"`
	Description string                        `json:"description"`
	Icon        string                        `json:"icon"`
	Categories  []string                      `json:"categories"`
	Localized   map[string]LocalizedStringsV1 `json:"localized"`
}

type LocalizedStringsV1 struct {
	Name        string `json:"name"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

type PackageV1 struct {
	VersionName string   `json:"versionName"`
	VersionCode int      `json:"versionCode"`
	Size        int64    `json:"size"`
	Hash        string   `json:"hash"`
	HashType    string   `json:"hashType"`
	APKName     string   `json:"apkName"`
	MinSDK      int      `json:"minSdkVersion"`
	TargetSDK   int      `json:"targetSdkVersion"`
	NativeCode  []string `json:"nativecode"`
}
