#include <stdint.h>
#include <pthread.h>
#include <stdlib.h>

typedef struct arg_struct {
    int aRowsBegin;
    int aRowsEnd;
    int aCols;
    int bCols;
    __uint128_t *a; 
    uint8_t *b; 
    __uint128_t *out;
} thr_args;

void multiply(int aRows, int aCols, int bCols, uint32_t *a, uint32_t *b, uint32_t *out); 

void binary_multiply(int aRows, int aCols, int bCols, uint32_t *a, uint8_t *b, uint32_t *out); 

void multiply128(int aRows, int aCols, int bCols, __uint128_t *a, __uint128_t *b, __uint128_t *out);

void binary_multiply128(int aRows, int aCols, int bCols, __uint128_t *a, uint8_t *b, __uint128_t *out);

void binary_multiply128_parallel(int aRowsBegin, int aRowsEnd, int aCols, int bCols, __uint128_t *a, uint8_t *b, __uint128_t *out);

void binary_multipyl128_parallel_bis(int nThreads, int rowsPerRoutine, int aRows, int aCols, int bCols, __uint128_t *a, uint8_t *b, __uint128_t *out);

void *mul_rows(void *arg);
