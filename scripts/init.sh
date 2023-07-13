#!/bin/bash

dirname=deploy
git clone https://$GITHUB_USER:$GITHUB_TOKEN@github.com/$GITHUB_ORG/$GITHUB_REPO.git $dirname

dirname=collect
git clone --branch result https://$GITHUB_USER:$GITHUB_TOKEN@github.com/$GITHUB_ORG/$GITHUB_REPO.git $dirname
