cd $(dirname $0)/src

cat ../.password.txt | sudo -S /etc/init.d/postgresql start

go run ./main.go $@
