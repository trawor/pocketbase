```
git fetch --tags github
MTAG=$(git tag --sort=-creatordate | head -n 1) && echo "'$MTAG' is target tag"

git checkout -b release $MTAG
git merge --no-edit --ff plugin/wechat
git merge --no-edit --ff plugin/wecom
git merge --no-edit --ff ui/plugin-settings
cd ui && npm run build && cd -
git add . && git commit -m "release $MTAG"
git tag -d $MTAG
git tag -a $MTAG -m "release $MTAG"
git push origin refs/tags/$MTAG
git checkout plugin/wechat
git branch -D release
```