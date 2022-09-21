#include <stdint.h>
#include <pthread.h>
#include <stdlib.h>

void multiply(int aRows, int aCols, int bCols, uint32_t *a, uint32_t *b, uint32_t *out); 

void binary_multiply(int aRows, int aCols, int bCols, uint32_t *a, uint8_t *b, uint32_t *out); 

void multiply128(int aRows, int aCols, int bCols, __uint128_t *a, __uint128_t *b, __uint128_t *out);

void binary_multiply128(int aRows, int aCols, int bCols, __uint128_t *a, uint8_t *b, __uint128_t *out);
