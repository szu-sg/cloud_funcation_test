cmake_minimum_required(VERSION 3.18)

project(C)

add_executable(main main.c)
target_compile_features(main PRIVATE cxx_std_17)