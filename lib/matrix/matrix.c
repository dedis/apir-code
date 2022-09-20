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

void multiply128(int aRows, int aCols, int bCols, __uint128_t *a, __uint128_t *b, __uint128_t *out) {
   	int i, j, k;
	for (i = 0; i < aRows; i++) {
		for (k = 0; k < aCols; k++) {
			for (j = 0; j < bCols; j++) {
				out[bCols*i+j] += a[aCols*i+k] * b[bCols*k+j];
			}
		}
	}
}

void binary_multiply128(int aRows, int aCols, int bCols, __uint128_t *a, uint8_t *b, __uint128_t *out) {
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

void binary_multipyl128_parallel_bis(int nThreads, int rowsPerRoutine, int aRows, int aCols, int bCols, __uint128_t *a, uint8_t *b, __uint128_t *out) {
	    int i;
		int begin, end;
		pthread_t *threads;
		threads = (pthread_t*)malloc(nThreads*sizeof(pthread_t));
		thr_args *args;
		args = malloc(sizeof(thr_args)*nThreads);
		for (i = 0; i < nThreads; i++) {
			begin = i*rowsPerRoutine;
			end = (i+1)*rowsPerRoutine;
			// make the last routine take all the left-over (from division) rows
			if (end > aRows) {
				end = aRows;
			}

			args[i] = (thr_args) {
				.aRowsBegin = begin,
				.aRowsEnd = end,
				.aCols = aCols,
				.bCols = bCols,
				.a = a,
				.b = b, 
				.out = out,
			};

			pthread_create(&threads[i], NULL, mul_rows, (void *) &args[i]);  
        }

		for (i = 0; i < nThreads; i++) {
      		//Joining all threads and collecting return value
      		pthread_join(threads[i], NULL);
    	}

		free(args);
		free(threads);
}

void binary_multiply128_parallel(int aRowsBegin, int aRowsEnd, int aCols, int bCols, __uint128_t *a, uint8_t *b, __uint128_t *out) {
   	int i, j, k;
	for (i = aRowsBegin; i < aRowsEnd; i++) {
		for (k = 0; k < aCols; k++) {
			for (j = 0; j < bCols; j++) {
				if (b[bCols*k+j] == 1) {
					out[bCols*i+j] += a[aCols*i+k];
				}
			}
		}
	}
}

void *mul_rows(void* arg) {
	struct arg_struct *arguments = arg;
	int i, j, k;
	for (i = arguments->aRowsBegin; i < arguments->aRowsEnd; i++) {
		for (k = 0; k < arguments->aCols; k++) {
			for (j = 0; j < arguments->bCols; j++) {
				if (arguments->b[arguments->bCols*k+j] == 1) {
					arguments->out[arguments->bCols*i+j] += arguments->a[arguments->aCols*i+k];
				}
			}
		}
	}

	pthread_exit(NULL);
    return NULL;
}
