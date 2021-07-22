# LICENSE:                                                                  
# This submission to NSS is to be made available under the terms of the
# Mozilla Public License, v. 2.0. You can obtain one at http:         
# //mozilla.org/MPL/2.0/. 
################################################################################
# Copyright(c) 2012, Intel Corp.

.set DATA, %xmm0
.set T, %xmm1
.set BSWAP_MASK, %xmm2
.set TMP0, %xmm3
.set TMP1, %xmm4
.set TMP2, %xmm5
.set TMP3, %xmm6
.set TMP4, %xmm7
.set Xhi, %xmm9

.Lpoly:
.quad 0x1, 0xc200000000000000 

#########################
# a = T
# b = TMP0 - remains unchanged
# res = T
# uses also TMP1,TMP2,TMP3,TMP4
# __m128i GFMUL(__m128i A, __m128i B);
.globl _GFMUL
#.type GFMUL,@function
_GFMUL:  
    vpclmulqdq  $0x00, TMP0, T, TMP1
    vpclmulqdq  $0x11, TMP0, T, TMP4

    vpshufd     $78, T, TMP2
    vpshufd     $78, TMP0, TMP3
    vpxor       T, TMP2, TMP2
    vpxor       TMP0, TMP3, TMP3

    vpclmulqdq  $0x00, TMP3, TMP2, TMP2
    vpxor       TMP1, TMP2, TMP2
    vpxor       TMP4, TMP2, TMP2

    vpslldq     $8, TMP2, TMP3
    vpsrldq     $8, TMP2, TMP2

    vpxor       TMP3, TMP1, TMP1
    vpxor       TMP2, TMP4, TMP4

    vpclmulqdq  $0x10, .Lpoly(%rip), TMP1, TMP2
    vpshufd     $78, TMP1, TMP3
    vpxor       TMP3, TMP2, TMP1

    vpclmulqdq  $0x10, .Lpoly(%rip), TMP1, TMP2
    vpshufd     $78, TMP1, TMP3
    vpxor       TMP3, TMP2, TMP1

    vpxor       TMP4, TMP1, T
    ret
#.size GFMUL, .-GFMUL
