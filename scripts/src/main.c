#include "some.h"

#include <stdio.h>

int main(){
    int a = 5;
    int b = 10;
    printf("The sum of %d and %d is %d\n", a, b, add(a, b));
    printf("The difference of %d and %d is %d\n", a, b, sub(a, b));
    return 0;
}
