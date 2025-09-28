# JavaSwitcher

⚠️ 開発にはClaude Codeを一部使用しています。
**高速Java環境切り替えツール** - 複数のJavaインストール間を瞬時に切り替える美しいCLIツール

[![Version](https://img.shields.io/badge/version-v0.2.4-blue.svg)](https://github.com/Sumire-Labs/JSwitcher)
[![Go](https://img.shields.io/badge/go-1.25+-00ADD8.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-OSL--3.0-green.svg)](LICENSE.md)

## 主要機能

### 高速切り替え
- **JAVA_HOME + PATH**: 環境変数の完全同期（永続化対応）
- **インスタント切り替え**: 矢印キー + Enterキーで瞬時に切り替え
- **自動検出**: システム内の全Javaインストールを自動発見

### 美しいUI
- **コンパクトモード**: 情報密度最適化（8-12行表示）
- **詳細モード**: 詳細情報表示
- **リアルタイム切り替え**: `c`キーでレイアウト変更

### フィルタリング
- **インクリメンタル検索**: `/`キーでリアルタイム絞り込み
- **履歴管理**: 最近使用したJava 5件を自動記録
- **プレビュー**: 選択中Javaの詳細情報表示

### クロスプラットフォーム
- **Windows**: レジストリ + PowerShell永続化
- **Linux/macOS**: シェル設定ファイル自動更新
- **完全永続化**: 再起動後も設定維持

## 対応Javaディストリビューション

- ✅ **Oracle JDK** - Commercial Java SE
- ✅ **OpenJDK** - Open source reference implementation
- ✅ **Eclipse Adoptium** - High-performance OpenJDK builds
- ✅ **Amazon Corretto** - No-cost, production-ready OpenJDK
- ✅ **Azul Zulu** - Enterprise-grade OpenJDK builds
- ✅ **GraalVM** - Universal virtual machine

## 📜 ライセンス

OSL-3.0 License - 詳細は [LICENSE.md](LICENSE.md) をご覧ください

---


