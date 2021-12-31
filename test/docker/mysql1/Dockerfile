FROM mysql:5.7

ENTRYPOINT ["mysql"]

CMD ["-h mysql", "-uroot", "-proot", "-e", "update test.user set name = 'one' where id = 1"]