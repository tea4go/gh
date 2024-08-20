# gh
# 生成版本标签
git tag -a v1.0.6 -m "Version 1.0.6"
git push origin --tags

## 设置临时代理
git config https.proxy https://127.0.0.1:32124
git config http.proxy  http://127.0.0.1:32124