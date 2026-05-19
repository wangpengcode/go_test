README 
This project for the beginner of the golang development, i am also news to golang who have 
years of java backend experience.
1. we need to launching the consul first for other service to registering in.
we use the consul for service discovering and instant configuration.

Use `./.tools/consul/consul agent -server -ui  -data-dir=./tools/data   -bootstrap-expect=1  -bind=127.0.0.1 -client=0.0.0.0`

This command will help us establish a consul service, you can access the ui web page by
   `http://127.0.0.1:8500/ui/`
you maybe config some key-value for your project.

2. launching the service of go_test_api
   ```you can just run with the main.go in the go_test_api```
It's easy to run this project.

3. launching the service of go_test_backend
   ```you can just run with the main.go in the go_test_backend```

It's easy to run this project.
4. other commands you may need
Use `go mod tidy`

5. add consul key-value configuration

![img.png](img.png)
6. you can request the black list 

![img_1.png](img_1.png)

7. other unit test for learning go lang in the package of concurrency

The directory of 'concurrency' maybe not suit for, i have some practice about, interface\
channel\locks, and things like that, you maybe add every knowledge about golang in this 
fileds with its unit test to demonstrate how golang works.