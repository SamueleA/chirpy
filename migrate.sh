#!/bin/bash

read -s -p "Enter password: " password

cd ./sql/schema && goose postgres "postgres://postgres:$password@localhost:5432/chirpy" $1