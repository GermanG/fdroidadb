// Copyright (C) 2026 German Gutierrez
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package fdroid

// IndexV1 structures (for backward compatibility)
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
	Signer      string   `json:"signer"`
}

// IndexV2 structures
type IndexV2 struct {
	Repo     RepoV2               `json:"repo"`
	Packages map[string]PackageV2 `json:"packages"`
}

type RepoV2 struct {
	Name        map[string]string          `json:"name"`
	Description map[string]string          `json:"description"`
	Icon        map[string]LocalizedFileV2 `json:"icon"`
	Timestamp   int64                      `json:"timestamp"`
}

type PackageV2 struct {
	Metadata MetadataV2            `json:"metadata"`
	Versions map[string]VersionV2 `json:"versions"`
}

type MetadataV2 struct {
	Name        map[string]string          `json:"name"`
	Summary     map[string]string          `json:"summary"`
	Description map[string]string          `json:"description"`
	Categories  []string                   `json:"categories"`
	Icon        map[string]LocalizedFileV2 `json:"icon"`
}

type LocalizedFileV2 struct {
	Name string `json:"name"`
}

type VersionV2 struct {
	File     FileV2     `json:"file"`
	Manifest ManifestV2 `json:"manifest"`
}

type ManifestV2 struct {
	VersionName string    `json:"versionName"`
	VersionCode int       `json:"versionCode"`
	NativeCode  []string  `json:"nativeCode"`
	Signer      *SignerV2 `json:"signer"`
	UsesSDK     *UsesSDKV2 `json:"usesSdk"`
}

type UsesSDKV2 struct {
	MinSDK    int `json:"minSdkVersion"`
	TargetSDK int `json:"targetSdkVersion"`
}

type SignerV2 struct {
	SHA256 []string `json:"sha256"`
}

type FileV2 struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// EntryV2 (from entry.json)
type EntryV2 struct {
	Index EntryIndexV2 `json:"index"`
}

type EntryIndexV2 struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}
