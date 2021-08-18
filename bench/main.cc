
#include <emmintrin.h>
#include <assert.h>
#include <iostream>
#include <xmmintrin.h>
#include "gfmul.h"
#include "dpf.h"

static const long logn = 10;
static const long rowlen = 1024*1024;

inline void xor_into(uint8_t *out, const uint8_t *other) {
  for (int i=0; i<rowlen; i++) {
    out[i] ^= other[i];
  }
}

inline void gcm_into(uint8_t *out, const uint8_t *mask, const uint8_t *row) {
  int chunks = rowlen / 16;
  __m128i a = _mm_load_si128((__m128i const*)&mask[0*16]);
  __m128i b = _mm_load_si128((__m128i const*)&row[0*16]);
  for (int c = 0; c < chunks; c++) {

    a = GFMUL(a, b);
   
    for (int i=0; i<8; i++) {
      uint16_t *p = (uint16_t*)&out[(c*16)+2*i];
      const int j = i;
      *p = _mm_extract_epi16(tmp, 0);
      *(p+2) = _mm_extract_epi16(tmp, 1);
      *(p+4) = _mm_extract_epi16(tmp, 2);
      *(p+6) = _mm_extract_epi16(tmp, 3);
      *(p+8) = _mm_extract_epi16(tmp, 4);
      *(p+10) = _mm_extract_epi16(tmp, 5);
      *(p+12) = _mm_extract_epi16(tmp, 6);
      *(p+14) = _mm_extract_epi16(tmp, 7);
    }
  }
}

void bench_dpf(std::vector<uint8_t> db) {
  auto pair = DPF::Gen(0, logn);

  clock_t start, end;

  // BEGIN TIMER
  start = clock();
  auto eval = DPF::EvalFull(pair.first, logn);

  std::vector<uint8_t> out(rowlen);
  for (int i = 0; i < (1<<(logn-3)); i++) {
    for (int j = 0; j < 8; j++) {
      if (eval[i] & (1 << j)) {
        xor_into(out.data(), db.data() + (rowlen*(8*i + j)));
      }
    }
  }
  
  // END TIMER
  end = clock();

  double elapsed = double(end - start) / double(CLOCKS_PER_SEC);
  double dbsize = double((1<<logn) * rowlen);// / double(1<<30);
  std::cout << "DB size: \t\t" << dbsize << std::endl;
  std::cout << "Elapsed: \t\t" << elapsed << std::endl;
  std::cout << "Classic DPF-based PIR\t\t";
  std::cout << "\t GiB/sec: " << dbsize/elapsed/double(1<<30) << std::endl;
}


void bench_gcm(std::vector<uint8_t> db) {
  auto pair = DPF::Gen(40, logn);

  clock_t start, end;

  // BEGIN TIMER
  start = clock();
  auto eval = DPF::EvalFull(pair.first, logn);

  std::vector<uint8_t> out(rowlen);
  for (int i = 0; i < (1<<logn); i++) {
    std::vector<uint8_t> rowbuf(rowlen);
    for (int j = 0; j < rowlen; j++) {
      rowbuf[i] = (256-i) & 0xFF;
    }

    gcm_into(out.data(), db.data() + (rowlen*i), rowbuf.data());
  }
  
  // END TIMER
  end = clock();

  double elapsed = double(end - start) / double(CLOCKS_PER_SEC);
  long long dbsize = (1<<logn) * rowlen;
  std::cout << "GCM-based PIR\t\t";
  std::cout << "\t GiB/sec: " << double(dbsize)/elapsed/(1 << 30) << std::endl;
}

int main(void) {
  assert(rowlen % 16 == 0);
  __m128i a = _mm_setr_epi32(1, 2, 3, 4);
  __m128i b = _mm_setr_epi32(1, 2, 3, 4);;

  __m128i c = GFMUL(a, b);

  std::vector<uint8_t> db((1<<logn) * rowlen);
  for (int i=0; i<db.size(); i++)  {
    db[i] = i & 0xFF;
  }

  std::cout << "Starting..." << std::endl;
  bench_dpf(db);
  bench_gcm(db);

  return 0;
}
