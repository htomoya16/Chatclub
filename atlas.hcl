env "local" {
    # 完成図（schema/配下の .sql/.hcl）
    src = "file://schema"
    dev = "docker://mysql/8.4/chatclub"
    
    # マイグレーション履歴の置き場
    migration {
        dir = "file://migrations"
        format = atlas
    }
    # 実際に適用する接続先
    url = "mysql://chatclub_user:chatclub_password@localhost:3306/chatclub"
}