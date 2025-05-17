# gh

# 生成版本标签

```
git tag -a v1.0.6 -m "Version 1.0.6"
git push origin --tags
```
# 提交代码

```
git status
git add .
git commit -m "auto::bugfix"
git push origin dev
```

# 设置临时代理

```
git config https.proxy https://127.0.0.1:32124
git config http.proxy  http://127.0.0.1:32124

git config --unset https.proxy
git config --unset http.proxy
```