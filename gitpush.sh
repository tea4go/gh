#!/bin/bash

cd /home/share/mycode/gitcode/gh

printf "=========================================================================\n"
echo 开始自动提交代码 ...... tea4go/gh
printf "=========================================================================\n"


# 检查是否有未暂存的修改
git diff --quiet
if [ $? -ne 0 ]; then
    # 检查是否有已暂存但未提交的修改
    git diff --cached --quiet
    if [ $? -ne 0 ]; then
        echo 没有代码修改，跳过提交。
        exit -1
    fi
fi

git status
git add .
git commit -m "auto::bugfix"
git push origin dev
git push
