#include "matrix.h"

void multiply(int aRows, int aCols, int bCols, uint32_t *a, uint32_t *b, uint32_t *out) {
   	int i, j, k;
	for (i = 0; i < aRows; i++) {
		for (k = 0; k < aCols; k++) {
			for (j = 0; j < bCols; j++) {
				out[bCols*i+j] += a[aCols*i+k] * b[bCols*k+j];
			}
		}
	}
}

void binary_multiply(int aRows, int aCols, int bCols, uint32_t *a, uint8_t *b, uint32_t *out) {
   	int i, j, k;
	for (i = 0; i < aRows; i++) {
		for (k = 0; k < aCols; k++) {
			for (j = 0; j < bCols; j++) {
				if (b[bCols*k+j] == 1) {
					out[bCols*i+j] += a[aCols*i+k];
				}
			}
		}
	}
}