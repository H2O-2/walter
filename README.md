Walter
======

<p align="center">
<img src="https://dl.dropboxusercontent.com/u/10177896/walter-logo-readme.png"/>
</p>

[![wercker status](https://app.wercker.com/status/4fcb4b110909fc45775d12641f5cf037/m "wercker status")](https://app.wercker.com/project/bykey/4fcb4b110909fc45775d12641f5cf037)

Walter is a tiny deployment pipeline tool.

----

Blogs
=====

* http://ainoya.io/walter
* http://walter-cd.net


----

Overview
========

Walter is a simple command line tool that automates build, test and deployment of applications or servers.

----

Getting Started
===============


How to install
--------------

Get a binary file from [GitHub Releases](https://github.com/walter-cd/walter/releases) and place it in `$PATH`.


Wrirte your pipeline
--------------------

Write your command pipeline in `pipeline.yml`.


```yaml
build:
  tasks:
    - name: setup build
      command: echo "setting up ..."
    - name: run build
      command: echo "building ..."
  cleanup:
    - name: cleanup build
      command: echo "cleanup build ..."

deploy:
  tasks:
    - name: run deploy
      command: echo "deploying ..."
  cleanup:
    - name: cleanup
      command: echo "cleanup deploy ..."
```


Run walter
----------

```
$ walter -build -deploy
INFO[0000] Build started
INFO[0000] [setup build] Start task
INFO[0000] [setup build] setting up ...
INFO[0000] [setup build] End task
INFO[0000] [run build] Start task
INFO[0000] [run build] building ...
INFO[0000] [run build] End task
INFO[0000] Build succeeded
INFO[0000] Build cleanup started
INFO[0000] [cleanup build] Start task
INFO[0000] [cleanup build] cleanup build ...
INFO[0000] [cleanup build] End task
INFO[0000] Build cleanup succeeded
INFO[0000] Deploy started
INFO[0000] [run deploy] Start task
INFO[0000] [run deploy] deploying ...
INFO[0000] [run deploy] End task
INFO[0000] Deploy succeeded
INFO[0000] [cleanup] Start task
INFO[0000] [cleanup] cleanup deploy ...
INFO[0000] [cleanup] End task
INFO[0000] Deploy cleanup succeeded
```

That's it.

----

Other features
==============

Environment variables
---------------------

You can use environment variables.

```yaml
deploy:
  tasks:
    - name: release files
      command: ghr -token $GITHUB_TOKEN $VERSION pkg/dist/$VERSION
```


Working directory
-----------------

You can specify a working directory of a task.

```yaml
build:
  tasks:
    - name: list files under /tmp
      command: ls
      directory: /tmp
```

Get stdout of a previous task
-----------------------------

Tasks get stdout of a previous task through a pipe.

```yaml
build:
  tasks:
    - name: setup build
      command: echo "setting up"
    - name: run build
      command: cat
```

The second "run build" task outputs "setting up".


Parallel tasks
--------------

You can define parallel tasks.

```yaml
build:
   tasks:
     - name: parallel tasks
       paralle:
           - name: task 1
             command: echo task 1
           - name: task 2
             command: echo task 2
           - name: task 3
             command: echo task 3
```

You can also mix serial tasks in parallel tasks.

```yaml
build:
   tasks:
     - name: parallel tasks
       parallel:
           - name: task 1
             command: echo task 1
           - name: task 2
             serial:
               - name: task 2-1
                 command: echo task 2-1
               - name: task 2-2
                 command: echo task 2-2
           - name: task 3
             command: echo task 3
```

   
Include
-------


Notification
------------

----

Changes in v2
=============



----

Contributing
============

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request

