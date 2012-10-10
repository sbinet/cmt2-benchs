cmt2-benchs
===========

``cmt2-benchs`` is a testbed to investigate various build systems
behaviours at ATLAS' scale.

Instructions
------------

``` sh
      
$ mkdir test
$ export CMTROOT=`pwd`
$ cd test

# generate 3 projects, each holding at most 8 packages with interdependencies
$ pyhon ../generator.py projects=3 packages=8

# generate a 'dot' file to visualize the dependency graph
$ dot -Tjpg generator.dot >generator.jpg

# run the CMake build
$ cd build
$ cmake --build=. ../CMakeLists.txt
$ cd ..
```
