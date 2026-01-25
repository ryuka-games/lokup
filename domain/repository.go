// Package domain はビジネスロジックの核心となるドメインモデルを定義する。
//
// このパッケージは外部依存を持たない。
// 他のパッケージ（features/, infrastructure/）はこのパッケージに依存するが、
// このパッケージは他に依存しない（Clean Architecture の依存ルール）。
package domain

// Repository は分析対象の GitHub リポジトリを表す値オブジェクト。
type Repository struct {
	Owner string // 例: "facebook"
	Name  string // 例: "react"
}

// FullName はリポジトリのフルネームを返す。
// 例: "facebook/react"
func (r Repository) FullName() string {
	return r.Owner + "/" + r.Name
}

// NewRepository は Repository を生成する。
func NewRepository(owner, name string) Repository {
	return Repository{
		Owner: owner,
		Name:  name,
	}
}
