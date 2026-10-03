# perf-large

<!-- 性能基准 fixture：1000 AC，由 perf/large.md 自动生成。6 种 EARS 句式轮换 + 数字响应。 -->

## User Stories

- As a 性能基准测试者 I want 一份 1000 AC 的大型 spec 文本 so that benchmark 测出大规模 lint 引擎压力。

## Acceptance Criteria

[AC-1] WHILE the matcher is hot on iteration 1 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-2] WHERE the cache is warm on iteration 2 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-3] UNLESS the input is empty on iteration 3 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-4] IF the engine sees event 4 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-5] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 5.
[AC-6] WHEN the bench harness runs iteration 6 THE SYSTEM SHALL complete within 5 ms.
[AC-7] WHILE the matcher is hot on iteration 7 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-8] WHERE the cache is warm on iteration 8 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-9] UNLESS the input is empty on iteration 9 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-10] IF the engine sees event 10 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-11] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 11.
[AC-12] WHEN the bench harness runs iteration 12 THE SYSTEM SHALL complete within 5 ms.
[AC-13] WHILE the matcher is hot on iteration 13 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-14] WHERE the cache is warm on iteration 14 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-15] UNLESS the input is empty on iteration 15 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-16] IF the engine sees event 16 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-17] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 17.
[AC-18] WHEN the bench harness runs iteration 18 THE SYSTEM SHALL complete within 5 ms.
[AC-19] WHILE the matcher is hot on iteration 19 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-20] WHERE the cache is warm on iteration 20 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-21] UNLESS the input is empty on iteration 21 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-22] IF the engine sees event 22 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-23] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 23.
[AC-24] WHEN the bench harness runs iteration 24 THE SYSTEM SHALL complete within 5 ms.
[AC-25] WHILE the matcher is hot on iteration 25 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-26] WHERE the cache is warm on iteration 26 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-27] UNLESS the input is empty on iteration 27 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-28] IF the engine sees event 28 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-29] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 29.
[AC-30] WHEN the bench harness runs iteration 30 THE SYSTEM SHALL complete within 5 ms.
[AC-31] WHILE the matcher is hot on iteration 31 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-32] WHERE the cache is warm on iteration 32 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-33] UNLESS the input is empty on iteration 33 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-34] IF the engine sees event 34 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-35] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 35.
[AC-36] WHEN the bench harness runs iteration 36 THE SYSTEM SHALL complete within 5 ms.
[AC-37] WHILE the matcher is hot on iteration 37 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-38] WHERE the cache is warm on iteration 38 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-39] UNLESS the input is empty on iteration 39 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-40] IF the engine sees event 40 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-41] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 41.
[AC-42] WHEN the bench harness runs iteration 42 THE SYSTEM SHALL complete within 5 ms.
[AC-43] WHILE the matcher is hot on iteration 43 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-44] WHERE the cache is warm on iteration 44 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-45] UNLESS the input is empty on iteration 45 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-46] IF the engine sees event 46 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-47] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 47.
[AC-48] WHEN the bench harness runs iteration 48 THE SYSTEM SHALL complete within 5 ms.
[AC-49] WHILE the matcher is hot on iteration 49 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-50] WHERE the cache is warm on iteration 50 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-51] UNLESS the input is empty on iteration 51 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-52] IF the engine sees event 52 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-53] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 53.
[AC-54] WHEN the bench harness runs iteration 54 THE SYSTEM SHALL complete within 5 ms.
[AC-55] WHILE the matcher is hot on iteration 55 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-56] WHERE the cache is warm on iteration 56 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-57] UNLESS the input is empty on iteration 57 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-58] IF the engine sees event 58 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-59] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 59.
[AC-60] WHEN the bench harness runs iteration 60 THE SYSTEM SHALL complete within 5 ms.
[AC-61] WHILE the matcher is hot on iteration 61 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-62] WHERE the cache is warm on iteration 62 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-63] UNLESS the input is empty on iteration 63 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-64] IF the engine sees event 64 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-65] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 65.
[AC-66] WHEN the bench harness runs iteration 66 THE SYSTEM SHALL complete within 5 ms.
[AC-67] WHILE the matcher is hot on iteration 67 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-68] WHERE the cache is warm on iteration 68 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-69] UNLESS the input is empty on iteration 69 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-70] IF the engine sees event 70 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-71] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 71.
[AC-72] WHEN the bench harness runs iteration 72 THE SYSTEM SHALL complete within 5 ms.
[AC-73] WHILE the matcher is hot on iteration 73 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-74] WHERE the cache is warm on iteration 74 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-75] UNLESS the input is empty on iteration 75 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-76] IF the engine sees event 76 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-77] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 77.
[AC-78] WHEN the bench harness runs iteration 78 THE SYSTEM SHALL complete within 5 ms.
[AC-79] WHILE the matcher is hot on iteration 79 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-80] WHERE the cache is warm on iteration 80 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-81] UNLESS the input is empty on iteration 81 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-82] IF the engine sees event 82 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-83] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 83.
[AC-84] WHEN the bench harness runs iteration 84 THE SYSTEM SHALL complete within 5 ms.
[AC-85] WHILE the matcher is hot on iteration 85 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-86] WHERE the cache is warm on iteration 86 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-87] UNLESS the input is empty on iteration 87 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-88] IF the engine sees event 88 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-89] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 89.
[AC-90] WHEN the bench harness runs iteration 90 THE SYSTEM SHALL complete within 5 ms.
[AC-91] WHILE the matcher is hot on iteration 91 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-92] WHERE the cache is warm on iteration 92 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-93] UNLESS the input is empty on iteration 93 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-94] IF the engine sees event 94 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-95] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 95.
[AC-96] WHEN the bench harness runs iteration 96 THE SYSTEM SHALL complete within 5 ms.
[AC-97] WHILE the matcher is hot on iteration 97 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-98] WHERE the cache is warm on iteration 98 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-99] UNLESS the input is empty on iteration 99 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-100] IF the engine sees event 100 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-101] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 101.
[AC-102] WHEN the bench harness runs iteration 102 THE SYSTEM SHALL complete within 5 ms.
[AC-103] WHILE the matcher is hot on iteration 103 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-104] WHERE the cache is warm on iteration 104 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-105] UNLESS the input is empty on iteration 105 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-106] IF the engine sees event 106 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-107] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 107.
[AC-108] WHEN the bench harness runs iteration 108 THE SYSTEM SHALL complete within 5 ms.
[AC-109] WHILE the matcher is hot on iteration 109 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-110] WHERE the cache is warm on iteration 110 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-111] UNLESS the input is empty on iteration 111 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-112] IF the engine sees event 112 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-113] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 113.
[AC-114] WHEN the bench harness runs iteration 114 THE SYSTEM SHALL complete within 5 ms.
[AC-115] WHILE the matcher is hot on iteration 115 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-116] WHERE the cache is warm on iteration 116 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-117] UNLESS the input is empty on iteration 117 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-118] IF the engine sees event 118 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-119] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 119.
[AC-120] WHEN the bench harness runs iteration 120 THE SYSTEM SHALL complete within 5 ms.
[AC-121] WHILE the matcher is hot on iteration 121 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-122] WHERE the cache is warm on iteration 122 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-123] UNLESS the input is empty on iteration 123 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-124] IF the engine sees event 124 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-125] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 125.
[AC-126] WHEN the bench harness runs iteration 126 THE SYSTEM SHALL complete within 5 ms.
[AC-127] WHILE the matcher is hot on iteration 127 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-128] WHERE the cache is warm on iteration 128 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-129] UNLESS the input is empty on iteration 129 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-130] IF the engine sees event 130 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-131] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 131.
[AC-132] WHEN the bench harness runs iteration 132 THE SYSTEM SHALL complete within 5 ms.
[AC-133] WHILE the matcher is hot on iteration 133 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-134] WHERE the cache is warm on iteration 134 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-135] UNLESS the input is empty on iteration 135 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-136] IF the engine sees event 136 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-137] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 137.
[AC-138] WHEN the bench harness runs iteration 138 THE SYSTEM SHALL complete within 5 ms.
[AC-139] WHILE the matcher is hot on iteration 139 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-140] WHERE the cache is warm on iteration 140 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-141] UNLESS the input is empty on iteration 141 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-142] IF the engine sees event 142 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-143] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 143.
[AC-144] WHEN the bench harness runs iteration 144 THE SYSTEM SHALL complete within 5 ms.
[AC-145] WHILE the matcher is hot on iteration 145 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-146] WHERE the cache is warm on iteration 146 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-147] UNLESS the input is empty on iteration 147 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-148] IF the engine sees event 148 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-149] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 149.
[AC-150] WHEN the bench harness runs iteration 150 THE SYSTEM SHALL complete within 5 ms.
[AC-151] WHILE the matcher is hot on iteration 151 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-152] WHERE the cache is warm on iteration 152 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-153] UNLESS the input is empty on iteration 153 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-154] IF the engine sees event 154 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-155] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 155.
[AC-156] WHEN the bench harness runs iteration 156 THE SYSTEM SHALL complete within 5 ms.
[AC-157] WHILE the matcher is hot on iteration 157 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-158] WHERE the cache is warm on iteration 158 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-159] UNLESS the input is empty on iteration 159 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-160] IF the engine sees event 160 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-161] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 161.
[AC-162] WHEN the bench harness runs iteration 162 THE SYSTEM SHALL complete within 5 ms.
[AC-163] WHILE the matcher is hot on iteration 163 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-164] WHERE the cache is warm on iteration 164 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-165] UNLESS the input is empty on iteration 165 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-166] IF the engine sees event 166 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-167] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 167.
[AC-168] WHEN the bench harness runs iteration 168 THE SYSTEM SHALL complete within 5 ms.
[AC-169] WHILE the matcher is hot on iteration 169 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-170] WHERE the cache is warm on iteration 170 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-171] UNLESS the input is empty on iteration 171 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-172] IF the engine sees event 172 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-173] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 173.
[AC-174] WHEN the bench harness runs iteration 174 THE SYSTEM SHALL complete within 5 ms.
[AC-175] WHILE the matcher is hot on iteration 175 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-176] WHERE the cache is warm on iteration 176 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-177] UNLESS the input is empty on iteration 177 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-178] IF the engine sees event 178 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-179] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 179.
[AC-180] WHEN the bench harness runs iteration 180 THE SYSTEM SHALL complete within 5 ms.
[AC-181] WHILE the matcher is hot on iteration 181 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-182] WHERE the cache is warm on iteration 182 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-183] UNLESS the input is empty on iteration 183 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-184] IF the engine sees event 184 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-185] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 185.
[AC-186] WHEN the bench harness runs iteration 186 THE SYSTEM SHALL complete within 5 ms.
[AC-187] WHILE the matcher is hot on iteration 187 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-188] WHERE the cache is warm on iteration 188 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-189] UNLESS the input is empty on iteration 189 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-190] IF the engine sees event 190 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-191] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 191.
[AC-192] WHEN the bench harness runs iteration 192 THE SYSTEM SHALL complete within 5 ms.
[AC-193] WHILE the matcher is hot on iteration 193 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-194] WHERE the cache is warm on iteration 194 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-195] UNLESS the input is empty on iteration 195 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-196] IF the engine sees event 196 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-197] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 197.
[AC-198] WHEN the bench harness runs iteration 198 THE SYSTEM SHALL complete within 5 ms.
[AC-199] WHILE the matcher is hot on iteration 199 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-200] WHERE the cache is warm on iteration 200 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-201] UNLESS the input is empty on iteration 201 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-202] IF the engine sees event 202 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-203] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 203.
[AC-204] WHEN the bench harness runs iteration 204 THE SYSTEM SHALL complete within 5 ms.
[AC-205] WHILE the matcher is hot on iteration 205 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-206] WHERE the cache is warm on iteration 206 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-207] UNLESS the input is empty on iteration 207 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-208] IF the engine sees event 208 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-209] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 209.
[AC-210] WHEN the bench harness runs iteration 210 THE SYSTEM SHALL complete within 5 ms.
[AC-211] WHILE the matcher is hot on iteration 211 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-212] WHERE the cache is warm on iteration 212 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-213] UNLESS the input is empty on iteration 213 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-214] IF the engine sees event 214 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-215] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 215.
[AC-216] WHEN the bench harness runs iteration 216 THE SYSTEM SHALL complete within 5 ms.
[AC-217] WHILE the matcher is hot on iteration 217 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-218] WHERE the cache is warm on iteration 218 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-219] UNLESS the input is empty on iteration 219 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-220] IF the engine sees event 220 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-221] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 221.
[AC-222] WHEN the bench harness runs iteration 222 THE SYSTEM SHALL complete within 5 ms.
[AC-223] WHILE the matcher is hot on iteration 223 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-224] WHERE the cache is warm on iteration 224 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-225] UNLESS the input is empty on iteration 225 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-226] IF the engine sees event 226 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-227] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 227.
[AC-228] WHEN the bench harness runs iteration 228 THE SYSTEM SHALL complete within 5 ms.
[AC-229] WHILE the matcher is hot on iteration 229 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-230] WHERE the cache is warm on iteration 230 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-231] UNLESS the input is empty on iteration 231 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-232] IF the engine sees event 232 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-233] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 233.
[AC-234] WHEN the bench harness runs iteration 234 THE SYSTEM SHALL complete within 5 ms.
[AC-235] WHILE the matcher is hot on iteration 235 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-236] WHERE the cache is warm on iteration 236 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-237] UNLESS the input is empty on iteration 237 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-238] IF the engine sees event 238 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-239] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 239.
[AC-240] WHEN the bench harness runs iteration 240 THE SYSTEM SHALL complete within 5 ms.
[AC-241] WHILE the matcher is hot on iteration 241 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-242] WHERE the cache is warm on iteration 242 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-243] UNLESS the input is empty on iteration 243 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-244] IF the engine sees event 244 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-245] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 245.
[AC-246] WHEN the bench harness runs iteration 246 THE SYSTEM SHALL complete within 5 ms.
[AC-247] WHILE the matcher is hot on iteration 247 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-248] WHERE the cache is warm on iteration 248 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-249] UNLESS the input is empty on iteration 249 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-250] IF the engine sees event 250 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-251] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 251.
[AC-252] WHEN the bench harness runs iteration 252 THE SYSTEM SHALL complete within 5 ms.
[AC-253] WHILE the matcher is hot on iteration 253 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-254] WHERE the cache is warm on iteration 254 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-255] UNLESS the input is empty on iteration 255 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-256] IF the engine sees event 256 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-257] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 257.
[AC-258] WHEN the bench harness runs iteration 258 THE SYSTEM SHALL complete within 5 ms.
[AC-259] WHILE the matcher is hot on iteration 259 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-260] WHERE the cache is warm on iteration 260 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-261] UNLESS the input is empty on iteration 261 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-262] IF the engine sees event 262 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-263] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 263.
[AC-264] WHEN the bench harness runs iteration 264 THE SYSTEM SHALL complete within 5 ms.
[AC-265] WHILE the matcher is hot on iteration 265 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-266] WHERE the cache is warm on iteration 266 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-267] UNLESS the input is empty on iteration 267 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-268] IF the engine sees event 268 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-269] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 269.
[AC-270] WHEN the bench harness runs iteration 270 THE SYSTEM SHALL complete within 5 ms.
[AC-271] WHILE the matcher is hot on iteration 271 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-272] WHERE the cache is warm on iteration 272 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-273] UNLESS the input is empty on iteration 273 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-274] IF the engine sees event 274 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-275] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 275.
[AC-276] WHEN the bench harness runs iteration 276 THE SYSTEM SHALL complete within 5 ms.
[AC-277] WHILE the matcher is hot on iteration 277 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-278] WHERE the cache is warm on iteration 278 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-279] UNLESS the input is empty on iteration 279 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-280] IF the engine sees event 280 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-281] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 281.
[AC-282] WHEN the bench harness runs iteration 282 THE SYSTEM SHALL complete within 5 ms.
[AC-283] WHILE the matcher is hot on iteration 283 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-284] WHERE the cache is warm on iteration 284 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-285] UNLESS the input is empty on iteration 285 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-286] IF the engine sees event 286 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-287] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 287.
[AC-288] WHEN the bench harness runs iteration 288 THE SYSTEM SHALL complete within 5 ms.
[AC-289] WHILE the matcher is hot on iteration 289 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-290] WHERE the cache is warm on iteration 290 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-291] UNLESS the input is empty on iteration 291 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-292] IF the engine sees event 292 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-293] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 293.
[AC-294] WHEN the bench harness runs iteration 294 THE SYSTEM SHALL complete within 5 ms.
[AC-295] WHILE the matcher is hot on iteration 295 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-296] WHERE the cache is warm on iteration 296 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-297] UNLESS the input is empty on iteration 297 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-298] IF the engine sees event 298 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-299] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 299.
[AC-300] WHEN the bench harness runs iteration 300 THE SYSTEM SHALL complete within 5 ms.
[AC-301] WHILE the matcher is hot on iteration 301 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-302] WHERE the cache is warm on iteration 302 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-303] UNLESS the input is empty on iteration 303 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-304] IF the engine sees event 304 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-305] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 305.
[AC-306] WHEN the bench harness runs iteration 306 THE SYSTEM SHALL complete within 5 ms.
[AC-307] WHILE the matcher is hot on iteration 307 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-308] WHERE the cache is warm on iteration 308 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-309] UNLESS the input is empty on iteration 309 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-310] IF the engine sees event 310 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-311] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 311.
[AC-312] WHEN the bench harness runs iteration 312 THE SYSTEM SHALL complete within 5 ms.
[AC-313] WHILE the matcher is hot on iteration 313 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-314] WHERE the cache is warm on iteration 314 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-315] UNLESS the input is empty on iteration 315 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-316] IF the engine sees event 316 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-317] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 317.
[AC-318] WHEN the bench harness runs iteration 318 THE SYSTEM SHALL complete within 5 ms.
[AC-319] WHILE the matcher is hot on iteration 319 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-320] WHERE the cache is warm on iteration 320 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-321] UNLESS the input is empty on iteration 321 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-322] IF the engine sees event 322 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-323] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 323.
[AC-324] WHEN the bench harness runs iteration 324 THE SYSTEM SHALL complete within 5 ms.
[AC-325] WHILE the matcher is hot on iteration 325 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-326] WHERE the cache is warm on iteration 326 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-327] UNLESS the input is empty on iteration 327 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-328] IF the engine sees event 328 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-329] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 329.
[AC-330] WHEN the bench harness runs iteration 330 THE SYSTEM SHALL complete within 5 ms.
[AC-331] WHILE the matcher is hot on iteration 331 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-332] WHERE the cache is warm on iteration 332 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-333] UNLESS the input is empty on iteration 333 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-334] IF the engine sees event 334 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-335] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 335.
[AC-336] WHEN the bench harness runs iteration 336 THE SYSTEM SHALL complete within 5 ms.
[AC-337] WHILE the matcher is hot on iteration 337 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-338] WHERE the cache is warm on iteration 338 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-339] UNLESS the input is empty on iteration 339 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-340] IF the engine sees event 340 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-341] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 341.
[AC-342] WHEN the bench harness runs iteration 342 THE SYSTEM SHALL complete within 5 ms.
[AC-343] WHILE the matcher is hot on iteration 343 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-344] WHERE the cache is warm on iteration 344 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-345] UNLESS the input is empty on iteration 345 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-346] IF the engine sees event 346 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-347] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 347.
[AC-348] WHEN the bench harness runs iteration 348 THE SYSTEM SHALL complete within 5 ms.
[AC-349] WHILE the matcher is hot on iteration 349 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-350] WHERE the cache is warm on iteration 350 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-351] UNLESS the input is empty on iteration 351 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-352] IF the engine sees event 352 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-353] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 353.
[AC-354] WHEN the bench harness runs iteration 354 THE SYSTEM SHALL complete within 5 ms.
[AC-355] WHILE the matcher is hot on iteration 355 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-356] WHERE the cache is warm on iteration 356 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-357] UNLESS the input is empty on iteration 357 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-358] IF the engine sees event 358 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-359] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 359.
[AC-360] WHEN the bench harness runs iteration 360 THE SYSTEM SHALL complete within 5 ms.
[AC-361] WHILE the matcher is hot on iteration 361 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-362] WHERE the cache is warm on iteration 362 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-363] UNLESS the input is empty on iteration 363 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-364] IF the engine sees event 364 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-365] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 365.
[AC-366] WHEN the bench harness runs iteration 366 THE SYSTEM SHALL complete within 5 ms.
[AC-367] WHILE the matcher is hot on iteration 367 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-368] WHERE the cache is warm on iteration 368 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-369] UNLESS the input is empty on iteration 369 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-370] IF the engine sees event 370 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-371] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 371.
[AC-372] WHEN the bench harness runs iteration 372 THE SYSTEM SHALL complete within 5 ms.
[AC-373] WHILE the matcher is hot on iteration 373 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-374] WHERE the cache is warm on iteration 374 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-375] UNLESS the input is empty on iteration 375 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-376] IF the engine sees event 376 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-377] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 377.
[AC-378] WHEN the bench harness runs iteration 378 THE SYSTEM SHALL complete within 5 ms.
[AC-379] WHILE the matcher is hot on iteration 379 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-380] WHERE the cache is warm on iteration 380 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-381] UNLESS the input is empty on iteration 381 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-382] IF the engine sees event 382 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-383] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 383.
[AC-384] WHEN the bench harness runs iteration 384 THE SYSTEM SHALL complete within 5 ms.
[AC-385] WHILE the matcher is hot on iteration 385 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-386] WHERE the cache is warm on iteration 386 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-387] UNLESS the input is empty on iteration 387 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-388] IF the engine sees event 388 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-389] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 389.
[AC-390] WHEN the bench harness runs iteration 390 THE SYSTEM SHALL complete within 5 ms.
[AC-391] WHILE the matcher is hot on iteration 391 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-392] WHERE the cache is warm on iteration 392 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-393] UNLESS the input is empty on iteration 393 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-394] IF the engine sees event 394 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-395] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 395.
[AC-396] WHEN the bench harness runs iteration 396 THE SYSTEM SHALL complete within 5 ms.
[AC-397] WHILE the matcher is hot on iteration 397 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-398] WHERE the cache is warm on iteration 398 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-399] UNLESS the input is empty on iteration 399 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-400] IF the engine sees event 400 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-401] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 401.
[AC-402] WHEN the bench harness runs iteration 402 THE SYSTEM SHALL complete within 5 ms.
[AC-403] WHILE the matcher is hot on iteration 403 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-404] WHERE the cache is warm on iteration 404 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-405] UNLESS the input is empty on iteration 405 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-406] IF the engine sees event 406 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-407] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 407.
[AC-408] WHEN the bench harness runs iteration 408 THE SYSTEM SHALL complete within 5 ms.
[AC-409] WHILE the matcher is hot on iteration 409 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-410] WHERE the cache is warm on iteration 410 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-411] UNLESS the input is empty on iteration 411 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-412] IF the engine sees event 412 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-413] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 413.
[AC-414] WHEN the bench harness runs iteration 414 THE SYSTEM SHALL complete within 5 ms.
[AC-415] WHILE the matcher is hot on iteration 415 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-416] WHERE the cache is warm on iteration 416 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-417] UNLESS the input is empty on iteration 417 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-418] IF the engine sees event 418 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-419] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 419.
[AC-420] WHEN the bench harness runs iteration 420 THE SYSTEM SHALL complete within 5 ms.
[AC-421] WHILE the matcher is hot on iteration 421 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-422] WHERE the cache is warm on iteration 422 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-423] UNLESS the input is empty on iteration 423 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-424] IF the engine sees event 424 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-425] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 425.
[AC-426] WHEN the bench harness runs iteration 426 THE SYSTEM SHALL complete within 5 ms.
[AC-427] WHILE the matcher is hot on iteration 427 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-428] WHERE the cache is warm on iteration 428 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-429] UNLESS the input is empty on iteration 429 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-430] IF the engine sees event 430 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-431] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 431.
[AC-432] WHEN the bench harness runs iteration 432 THE SYSTEM SHALL complete within 5 ms.
[AC-433] WHILE the matcher is hot on iteration 433 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-434] WHERE the cache is warm on iteration 434 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-435] UNLESS the input is empty on iteration 435 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-436] IF the engine sees event 436 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-437] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 437.
[AC-438] WHEN the bench harness runs iteration 438 THE SYSTEM SHALL complete within 5 ms.
[AC-439] WHILE the matcher is hot on iteration 439 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-440] WHERE the cache is warm on iteration 440 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-441] UNLESS the input is empty on iteration 441 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-442] IF the engine sees event 442 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-443] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 443.
[AC-444] WHEN the bench harness runs iteration 444 THE SYSTEM SHALL complete within 5 ms.
[AC-445] WHILE the matcher is hot on iteration 445 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-446] WHERE the cache is warm on iteration 446 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-447] UNLESS the input is empty on iteration 447 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-448] IF the engine sees event 448 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-449] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 449.
[AC-450] WHEN the bench harness runs iteration 450 THE SYSTEM SHALL complete within 5 ms.
[AC-451] WHILE the matcher is hot on iteration 451 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-452] WHERE the cache is warm on iteration 452 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-453] UNLESS the input is empty on iteration 453 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-454] IF the engine sees event 454 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-455] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 455.
[AC-456] WHEN the bench harness runs iteration 456 THE SYSTEM SHALL complete within 5 ms.
[AC-457] WHILE the matcher is hot on iteration 457 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-458] WHERE the cache is warm on iteration 458 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-459] UNLESS the input is empty on iteration 459 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-460] IF the engine sees event 460 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-461] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 461.
[AC-462] WHEN the bench harness runs iteration 462 THE SYSTEM SHALL complete within 5 ms.
[AC-463] WHILE the matcher is hot on iteration 463 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-464] WHERE the cache is warm on iteration 464 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-465] UNLESS the input is empty on iteration 465 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-466] IF the engine sees event 466 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-467] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 467.
[AC-468] WHEN the bench harness runs iteration 468 THE SYSTEM SHALL complete within 5 ms.
[AC-469] WHILE the matcher is hot on iteration 469 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-470] WHERE the cache is warm on iteration 470 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-471] UNLESS the input is empty on iteration 471 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-472] IF the engine sees event 472 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-473] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 473.
[AC-474] WHEN the bench harness runs iteration 474 THE SYSTEM SHALL complete within 5 ms.
[AC-475] WHILE the matcher is hot on iteration 475 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-476] WHERE the cache is warm on iteration 476 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-477] UNLESS the input is empty on iteration 477 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-478] IF the engine sees event 478 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-479] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 479.
[AC-480] WHEN the bench harness runs iteration 480 THE SYSTEM SHALL complete within 5 ms.
[AC-481] WHILE the matcher is hot on iteration 481 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-482] WHERE the cache is warm on iteration 482 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-483] UNLESS the input is empty on iteration 483 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-484] IF the engine sees event 484 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-485] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 485.
[AC-486] WHEN the bench harness runs iteration 486 THE SYSTEM SHALL complete within 5 ms.
[AC-487] WHILE the matcher is hot on iteration 487 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-488] WHERE the cache is warm on iteration 488 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-489] UNLESS the input is empty on iteration 489 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-490] IF the engine sees event 490 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-491] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 491.
[AC-492] WHEN the bench harness runs iteration 492 THE SYSTEM SHALL complete within 5 ms.
[AC-493] WHILE the matcher is hot on iteration 493 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-494] WHERE the cache is warm on iteration 494 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-495] UNLESS the input is empty on iteration 495 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-496] IF the engine sees event 496 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-497] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 497.
[AC-498] WHEN the bench harness runs iteration 498 THE SYSTEM SHALL complete within 5 ms.
[AC-499] WHILE the matcher is hot on iteration 499 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-500] WHERE the cache is warm on iteration 500 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-501] UNLESS the input is empty on iteration 501 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-502] IF the engine sees event 502 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-503] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 503.
[AC-504] WHEN the bench harness runs iteration 504 THE SYSTEM SHALL complete within 5 ms.
[AC-505] WHILE the matcher is hot on iteration 505 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-506] WHERE the cache is warm on iteration 506 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-507] UNLESS the input is empty on iteration 507 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-508] IF the engine sees event 508 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-509] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 509.
[AC-510] WHEN the bench harness runs iteration 510 THE SYSTEM SHALL complete within 5 ms.
[AC-511] WHILE the matcher is hot on iteration 511 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-512] WHERE the cache is warm on iteration 512 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-513] UNLESS the input is empty on iteration 513 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-514] IF the engine sees event 514 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-515] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 515.
[AC-516] WHEN the bench harness runs iteration 516 THE SYSTEM SHALL complete within 5 ms.
[AC-517] WHILE the matcher is hot on iteration 517 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-518] WHERE the cache is warm on iteration 518 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-519] UNLESS the input is empty on iteration 519 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-520] IF the engine sees event 520 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-521] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 521.
[AC-522] WHEN the bench harness runs iteration 522 THE SYSTEM SHALL complete within 5 ms.
[AC-523] WHILE the matcher is hot on iteration 523 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-524] WHERE the cache is warm on iteration 524 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-525] UNLESS the input is empty on iteration 525 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-526] IF the engine sees event 526 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-527] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 527.
[AC-528] WHEN the bench harness runs iteration 528 THE SYSTEM SHALL complete within 5 ms.
[AC-529] WHILE the matcher is hot on iteration 529 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-530] WHERE the cache is warm on iteration 530 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-531] UNLESS the input is empty on iteration 531 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-532] IF the engine sees event 532 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-533] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 533.
[AC-534] WHEN the bench harness runs iteration 534 THE SYSTEM SHALL complete within 5 ms.
[AC-535] WHILE the matcher is hot on iteration 535 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-536] WHERE the cache is warm on iteration 536 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-537] UNLESS the input is empty on iteration 537 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-538] IF the engine sees event 538 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-539] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 539.
[AC-540] WHEN the bench harness runs iteration 540 THE SYSTEM SHALL complete within 5 ms.
[AC-541] WHILE the matcher is hot on iteration 541 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-542] WHERE the cache is warm on iteration 542 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-543] UNLESS the input is empty on iteration 543 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-544] IF the engine sees event 544 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-545] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 545.
[AC-546] WHEN the bench harness runs iteration 546 THE SYSTEM SHALL complete within 5 ms.
[AC-547] WHILE the matcher is hot on iteration 547 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-548] WHERE the cache is warm on iteration 548 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-549] UNLESS the input is empty on iteration 549 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-550] IF the engine sees event 550 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-551] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 551.
[AC-552] WHEN the bench harness runs iteration 552 THE SYSTEM SHALL complete within 5 ms.
[AC-553] WHILE the matcher is hot on iteration 553 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-554] WHERE the cache is warm on iteration 554 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-555] UNLESS the input is empty on iteration 555 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-556] IF the engine sees event 556 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-557] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 557.
[AC-558] WHEN the bench harness runs iteration 558 THE SYSTEM SHALL complete within 5 ms.
[AC-559] WHILE the matcher is hot on iteration 559 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-560] WHERE the cache is warm on iteration 560 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-561] UNLESS the input is empty on iteration 561 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-562] IF the engine sees event 562 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-563] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 563.
[AC-564] WHEN the bench harness runs iteration 564 THE SYSTEM SHALL complete within 5 ms.
[AC-565] WHILE the matcher is hot on iteration 565 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-566] WHERE the cache is warm on iteration 566 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-567] UNLESS the input is empty on iteration 567 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-568] IF the engine sees event 568 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-569] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 569.
[AC-570] WHEN the bench harness runs iteration 570 THE SYSTEM SHALL complete within 5 ms.
[AC-571] WHILE the matcher is hot on iteration 571 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-572] WHERE the cache is warm on iteration 572 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-573] UNLESS the input is empty on iteration 573 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-574] IF the engine sees event 574 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-575] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 575.
[AC-576] WHEN the bench harness runs iteration 576 THE SYSTEM SHALL complete within 5 ms.
[AC-577] WHILE the matcher is hot on iteration 577 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-578] WHERE the cache is warm on iteration 578 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-579] UNLESS the input is empty on iteration 579 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-580] IF the engine sees event 580 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-581] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 581.
[AC-582] WHEN the bench harness runs iteration 582 THE SYSTEM SHALL complete within 5 ms.
[AC-583] WHILE the matcher is hot on iteration 583 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-584] WHERE the cache is warm on iteration 584 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-585] UNLESS the input is empty on iteration 585 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-586] IF the engine sees event 586 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-587] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 587.
[AC-588] WHEN the bench harness runs iteration 588 THE SYSTEM SHALL complete within 5 ms.
[AC-589] WHILE the matcher is hot on iteration 589 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-590] WHERE the cache is warm on iteration 590 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-591] UNLESS the input is empty on iteration 591 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-592] IF the engine sees event 592 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-593] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 593.
[AC-594] WHEN the bench harness runs iteration 594 THE SYSTEM SHALL complete within 5 ms.
[AC-595] WHILE the matcher is hot on iteration 595 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-596] WHERE the cache is warm on iteration 596 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-597] UNLESS the input is empty on iteration 597 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-598] IF the engine sees event 598 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-599] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 599.
[AC-600] WHEN the bench harness runs iteration 600 THE SYSTEM SHALL complete within 5 ms.
[AC-601] WHILE the matcher is hot on iteration 601 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-602] WHERE the cache is warm on iteration 602 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-603] UNLESS the input is empty on iteration 603 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-604] IF the engine sees event 604 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-605] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 605.
[AC-606] WHEN the bench harness runs iteration 606 THE SYSTEM SHALL complete within 5 ms.
[AC-607] WHILE the matcher is hot on iteration 607 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-608] WHERE the cache is warm on iteration 608 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-609] UNLESS the input is empty on iteration 609 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-610] IF the engine sees event 610 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-611] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 611.
[AC-612] WHEN the bench harness runs iteration 612 THE SYSTEM SHALL complete within 5 ms.
[AC-613] WHILE the matcher is hot on iteration 613 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-614] WHERE the cache is warm on iteration 614 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-615] UNLESS the input is empty on iteration 615 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-616] IF the engine sees event 616 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-617] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 617.
[AC-618] WHEN the bench harness runs iteration 618 THE SYSTEM SHALL complete within 5 ms.
[AC-619] WHILE the matcher is hot on iteration 619 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-620] WHERE the cache is warm on iteration 620 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-621] UNLESS the input is empty on iteration 621 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-622] IF the engine sees event 622 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-623] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 623.
[AC-624] WHEN the bench harness runs iteration 624 THE SYSTEM SHALL complete within 5 ms.
[AC-625] WHILE the matcher is hot on iteration 625 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-626] WHERE the cache is warm on iteration 626 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-627] UNLESS the input is empty on iteration 627 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-628] IF the engine sees event 628 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-629] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 629.
[AC-630] WHEN the bench harness runs iteration 630 THE SYSTEM SHALL complete within 5 ms.
[AC-631] WHILE the matcher is hot on iteration 631 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-632] WHERE the cache is warm on iteration 632 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-633] UNLESS the input is empty on iteration 633 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-634] IF the engine sees event 634 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-635] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 635.
[AC-636] WHEN the bench harness runs iteration 636 THE SYSTEM SHALL complete within 5 ms.
[AC-637] WHILE the matcher is hot on iteration 637 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-638] WHERE the cache is warm on iteration 638 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-639] UNLESS the input is empty on iteration 639 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-640] IF the engine sees event 640 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-641] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 641.
[AC-642] WHEN the bench harness runs iteration 642 THE SYSTEM SHALL complete within 5 ms.
[AC-643] WHILE the matcher is hot on iteration 643 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-644] WHERE the cache is warm on iteration 644 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-645] UNLESS the input is empty on iteration 645 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-646] IF the engine sees event 646 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-647] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 647.
[AC-648] WHEN the bench harness runs iteration 648 THE SYSTEM SHALL complete within 5 ms.
[AC-649] WHILE the matcher is hot on iteration 649 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-650] WHERE the cache is warm on iteration 650 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-651] UNLESS the input is empty on iteration 651 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-652] IF the engine sees event 652 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-653] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 653.
[AC-654] WHEN the bench harness runs iteration 654 THE SYSTEM SHALL complete within 5 ms.
[AC-655] WHILE the matcher is hot on iteration 655 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-656] WHERE the cache is warm on iteration 656 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-657] UNLESS the input is empty on iteration 657 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-658] IF the engine sees event 658 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-659] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 659.
[AC-660] WHEN the bench harness runs iteration 660 THE SYSTEM SHALL complete within 5 ms.
[AC-661] WHILE the matcher is hot on iteration 661 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-662] WHERE the cache is warm on iteration 662 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-663] UNLESS the input is empty on iteration 663 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-664] IF the engine sees event 664 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-665] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 665.
[AC-666] WHEN the bench harness runs iteration 666 THE SYSTEM SHALL complete within 5 ms.
[AC-667] WHILE the matcher is hot on iteration 667 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-668] WHERE the cache is warm on iteration 668 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-669] UNLESS the input is empty on iteration 669 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-670] IF the engine sees event 670 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-671] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 671.
[AC-672] WHEN the bench harness runs iteration 672 THE SYSTEM SHALL complete within 5 ms.
[AC-673] WHILE the matcher is hot on iteration 673 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-674] WHERE the cache is warm on iteration 674 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-675] UNLESS the input is empty on iteration 675 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-676] IF the engine sees event 676 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-677] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 677.
[AC-678] WHEN the bench harness runs iteration 678 THE SYSTEM SHALL complete within 5 ms.
[AC-679] WHILE the matcher is hot on iteration 679 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-680] WHERE the cache is warm on iteration 680 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-681] UNLESS the input is empty on iteration 681 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-682] IF the engine sees event 682 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-683] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 683.
[AC-684] WHEN the bench harness runs iteration 684 THE SYSTEM SHALL complete within 5 ms.
[AC-685] WHILE the matcher is hot on iteration 685 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-686] WHERE the cache is warm on iteration 686 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-687] UNLESS the input is empty on iteration 687 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-688] IF the engine sees event 688 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-689] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 689.
[AC-690] WHEN the bench harness runs iteration 690 THE SYSTEM SHALL complete within 5 ms.
[AC-691] WHILE the matcher is hot on iteration 691 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-692] WHERE the cache is warm on iteration 692 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-693] UNLESS the input is empty on iteration 693 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-694] IF the engine sees event 694 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-695] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 695.
[AC-696] WHEN the bench harness runs iteration 696 THE SYSTEM SHALL complete within 5 ms.
[AC-697] WHILE the matcher is hot on iteration 697 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-698] WHERE the cache is warm on iteration 698 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-699] UNLESS the input is empty on iteration 699 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-700] IF the engine sees event 700 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-701] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 701.
[AC-702] WHEN the bench harness runs iteration 702 THE SYSTEM SHALL complete within 5 ms.
[AC-703] WHILE the matcher is hot on iteration 703 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-704] WHERE the cache is warm on iteration 704 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-705] UNLESS the input is empty on iteration 705 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-706] IF the engine sees event 706 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-707] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 707.
[AC-708] WHEN the bench harness runs iteration 708 THE SYSTEM SHALL complete within 5 ms.
[AC-709] WHILE the matcher is hot on iteration 709 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-710] WHERE the cache is warm on iteration 710 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-711] UNLESS the input is empty on iteration 711 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-712] IF the engine sees event 712 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-713] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 713.
[AC-714] WHEN the bench harness runs iteration 714 THE SYSTEM SHALL complete within 5 ms.
[AC-715] WHILE the matcher is hot on iteration 715 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-716] WHERE the cache is warm on iteration 716 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-717] UNLESS the input is empty on iteration 717 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-718] IF the engine sees event 718 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-719] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 719.
[AC-720] WHEN the bench harness runs iteration 720 THE SYSTEM SHALL complete within 5 ms.
[AC-721] WHILE the matcher is hot on iteration 721 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-722] WHERE the cache is warm on iteration 722 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-723] UNLESS the input is empty on iteration 723 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-724] IF the engine sees event 724 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-725] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 725.
[AC-726] WHEN the bench harness runs iteration 726 THE SYSTEM SHALL complete within 5 ms.
[AC-727] WHILE the matcher is hot on iteration 727 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-728] WHERE the cache is warm on iteration 728 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-729] UNLESS the input is empty on iteration 729 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-730] IF the engine sees event 730 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-731] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 731.
[AC-732] WHEN the bench harness runs iteration 732 THE SYSTEM SHALL complete within 5 ms.
[AC-733] WHILE the matcher is hot on iteration 733 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-734] WHERE the cache is warm on iteration 734 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-735] UNLESS the input is empty on iteration 735 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-736] IF the engine sees event 736 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-737] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 737.
[AC-738] WHEN the bench harness runs iteration 738 THE SYSTEM SHALL complete within 5 ms.
[AC-739] WHILE the matcher is hot on iteration 739 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-740] WHERE the cache is warm on iteration 740 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-741] UNLESS the input is empty on iteration 741 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-742] IF the engine sees event 742 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-743] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 743.
[AC-744] WHEN the bench harness runs iteration 744 THE SYSTEM SHALL complete within 5 ms.
[AC-745] WHILE the matcher is hot on iteration 745 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-746] WHERE the cache is warm on iteration 746 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-747] UNLESS the input is empty on iteration 747 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-748] IF the engine sees event 748 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-749] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 749.
[AC-750] WHEN the bench harness runs iteration 750 THE SYSTEM SHALL complete within 5 ms.
[AC-751] WHILE the matcher is hot on iteration 751 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-752] WHERE the cache is warm on iteration 752 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-753] UNLESS the input is empty on iteration 753 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-754] IF the engine sees event 754 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-755] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 755.
[AC-756] WHEN the bench harness runs iteration 756 THE SYSTEM SHALL complete within 5 ms.
[AC-757] WHILE the matcher is hot on iteration 757 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-758] WHERE the cache is warm on iteration 758 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-759] UNLESS the input is empty on iteration 759 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-760] IF the engine sees event 760 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-761] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 761.
[AC-762] WHEN the bench harness runs iteration 762 THE SYSTEM SHALL complete within 5 ms.
[AC-763] WHILE the matcher is hot on iteration 763 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-764] WHERE the cache is warm on iteration 764 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-765] UNLESS the input is empty on iteration 765 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-766] IF the engine sees event 766 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-767] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 767.
[AC-768] WHEN the bench harness runs iteration 768 THE SYSTEM SHALL complete within 5 ms.
[AC-769] WHILE the matcher is hot on iteration 769 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-770] WHERE the cache is warm on iteration 770 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-771] UNLESS the input is empty on iteration 771 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-772] IF the engine sees event 772 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-773] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 773.
[AC-774] WHEN the bench harness runs iteration 774 THE SYSTEM SHALL complete within 5 ms.
[AC-775] WHILE the matcher is hot on iteration 775 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-776] WHERE the cache is warm on iteration 776 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-777] UNLESS the input is empty on iteration 777 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-778] IF the engine sees event 778 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-779] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 779.
[AC-780] WHEN the bench harness runs iteration 780 THE SYSTEM SHALL complete within 5 ms.
[AC-781] WHILE the matcher is hot on iteration 781 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-782] WHERE the cache is warm on iteration 782 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-783] UNLESS the input is empty on iteration 783 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-784] IF the engine sees event 784 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-785] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 785.
[AC-786] WHEN the bench harness runs iteration 786 THE SYSTEM SHALL complete within 5 ms.
[AC-787] WHILE the matcher is hot on iteration 787 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-788] WHERE the cache is warm on iteration 788 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-789] UNLESS the input is empty on iteration 789 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-790] IF the engine sees event 790 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-791] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 791.
[AC-792] WHEN the bench harness runs iteration 792 THE SYSTEM SHALL complete within 5 ms.
[AC-793] WHILE the matcher is hot on iteration 793 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-794] WHERE the cache is warm on iteration 794 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-795] UNLESS the input is empty on iteration 795 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-796] IF the engine sees event 796 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-797] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 797.
[AC-798] WHEN the bench harness runs iteration 798 THE SYSTEM SHALL complete within 5 ms.
[AC-799] WHILE the matcher is hot on iteration 799 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-800] WHERE the cache is warm on iteration 800 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-801] UNLESS the input is empty on iteration 801 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-802] IF the engine sees event 802 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-803] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 803.
[AC-804] WHEN the bench harness runs iteration 804 THE SYSTEM SHALL complete within 5 ms.
[AC-805] WHILE the matcher is hot on iteration 805 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-806] WHERE the cache is warm on iteration 806 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-807] UNLESS the input is empty on iteration 807 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-808] IF the engine sees event 808 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-809] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 809.
[AC-810] WHEN the bench harness runs iteration 810 THE SYSTEM SHALL complete within 5 ms.
[AC-811] WHILE the matcher is hot on iteration 811 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-812] WHERE the cache is warm on iteration 812 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-813] UNLESS the input is empty on iteration 813 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-814] IF the engine sees event 814 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-815] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 815.
[AC-816] WHEN the bench harness runs iteration 816 THE SYSTEM SHALL complete within 5 ms.
[AC-817] WHILE the matcher is hot on iteration 817 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-818] WHERE the cache is warm on iteration 818 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-819] UNLESS the input is empty on iteration 819 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-820] IF the engine sees event 820 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-821] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 821.
[AC-822] WHEN the bench harness runs iteration 822 THE SYSTEM SHALL complete within 5 ms.
[AC-823] WHILE the matcher is hot on iteration 823 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-824] WHERE the cache is warm on iteration 824 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-825] UNLESS the input is empty on iteration 825 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-826] IF the engine sees event 826 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-827] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 827.
[AC-828] WHEN the bench harness runs iteration 828 THE SYSTEM SHALL complete within 5 ms.
[AC-829] WHILE the matcher is hot on iteration 829 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-830] WHERE the cache is warm on iteration 830 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-831] UNLESS the input is empty on iteration 831 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-832] IF the engine sees event 832 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-833] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 833.
[AC-834] WHEN the bench harness runs iteration 834 THE SYSTEM SHALL complete within 5 ms.
[AC-835] WHILE the matcher is hot on iteration 835 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-836] WHERE the cache is warm on iteration 836 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-837] UNLESS the input is empty on iteration 837 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-838] IF the engine sees event 838 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-839] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 839.
[AC-840] WHEN the bench harness runs iteration 840 THE SYSTEM SHALL complete within 5 ms.
[AC-841] WHILE the matcher is hot on iteration 841 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-842] WHERE the cache is warm on iteration 842 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-843] UNLESS the input is empty on iteration 843 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-844] IF the engine sees event 844 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-845] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 845.
[AC-846] WHEN the bench harness runs iteration 846 THE SYSTEM SHALL complete within 5 ms.
[AC-847] WHILE the matcher is hot on iteration 847 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-848] WHERE the cache is warm on iteration 848 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-849] UNLESS the input is empty on iteration 849 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-850] IF the engine sees event 850 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-851] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 851.
[AC-852] WHEN the bench harness runs iteration 852 THE SYSTEM SHALL complete within 5 ms.
[AC-853] WHILE the matcher is hot on iteration 853 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-854] WHERE the cache is warm on iteration 854 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-855] UNLESS the input is empty on iteration 855 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-856] IF the engine sees event 856 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-857] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 857.
[AC-858] WHEN the bench harness runs iteration 858 THE SYSTEM SHALL complete within 5 ms.
[AC-859] WHILE the matcher is hot on iteration 859 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-860] WHERE the cache is warm on iteration 860 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-861] UNLESS the input is empty on iteration 861 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-862] IF the engine sees event 862 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-863] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 863.
[AC-864] WHEN the bench harness runs iteration 864 THE SYSTEM SHALL complete within 5 ms.
[AC-865] WHILE the matcher is hot on iteration 865 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-866] WHERE the cache is warm on iteration 866 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-867] UNLESS the input is empty on iteration 867 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-868] IF the engine sees event 868 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-869] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 869.
[AC-870] WHEN the bench harness runs iteration 870 THE SYSTEM SHALL complete within 5 ms.
[AC-871] WHILE the matcher is hot on iteration 871 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-872] WHERE the cache is warm on iteration 872 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-873] UNLESS the input is empty on iteration 873 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-874] IF the engine sees event 874 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-875] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 875.
[AC-876] WHEN the bench harness runs iteration 876 THE SYSTEM SHALL complete within 5 ms.
[AC-877] WHILE the matcher is hot on iteration 877 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-878] WHERE the cache is warm on iteration 878 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-879] UNLESS the input is empty on iteration 879 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-880] IF the engine sees event 880 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-881] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 881.
[AC-882] WHEN the bench harness runs iteration 882 THE SYSTEM SHALL complete within 5 ms.
[AC-883] WHILE the matcher is hot on iteration 883 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-884] WHERE the cache is warm on iteration 884 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-885] UNLESS the input is empty on iteration 885 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-886] IF the engine sees event 886 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-887] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 887.
[AC-888] WHEN the bench harness runs iteration 888 THE SYSTEM SHALL complete within 5 ms.
[AC-889] WHILE the matcher is hot on iteration 889 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-890] WHERE the cache is warm on iteration 890 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-891] UNLESS the input is empty on iteration 891 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-892] IF the engine sees event 892 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-893] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 893.
[AC-894] WHEN the bench harness runs iteration 894 THE SYSTEM SHALL complete within 5 ms.
[AC-895] WHILE the matcher is hot on iteration 895 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-896] WHERE the cache is warm on iteration 896 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-897] UNLESS the input is empty on iteration 897 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-898] IF the engine sees event 898 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-899] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 899.
[AC-900] WHEN the bench harness runs iteration 900 THE SYSTEM SHALL complete within 5 ms.
[AC-901] WHILE the matcher is hot on iteration 901 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-902] WHERE the cache is warm on iteration 902 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-903] UNLESS the input is empty on iteration 903 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-904] IF the engine sees event 904 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-905] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 905.
[AC-906] WHEN the bench harness runs iteration 906 THE SYSTEM SHALL complete within 5 ms.
[AC-907] WHILE the matcher is hot on iteration 907 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-908] WHERE the cache is warm on iteration 908 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-909] UNLESS the input is empty on iteration 909 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-910] IF the engine sees event 910 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-911] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 911.
[AC-912] WHEN the bench harness runs iteration 912 THE SYSTEM SHALL complete within 5 ms.
[AC-913] WHILE the matcher is hot on iteration 913 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-914] WHERE the cache is warm on iteration 914 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-915] UNLESS the input is empty on iteration 915 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-916] IF the engine sees event 916 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-917] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 917.
[AC-918] WHEN the bench harness runs iteration 918 THE SYSTEM SHALL complete within 5 ms.
[AC-919] WHILE the matcher is hot on iteration 919 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-920] WHERE the cache is warm on iteration 920 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-921] UNLESS the input is empty on iteration 921 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-922] IF the engine sees event 922 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-923] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 923.
[AC-924] WHEN the bench harness runs iteration 924 THE SYSTEM SHALL complete within 5 ms.
[AC-925] WHILE the matcher is hot on iteration 925 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-926] WHERE the cache is warm on iteration 926 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-927] UNLESS the input is empty on iteration 927 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-928] IF the engine sees event 928 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-929] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 929.
[AC-930] WHEN the bench harness runs iteration 930 THE SYSTEM SHALL complete within 5 ms.
[AC-931] WHILE the matcher is hot on iteration 931 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-932] WHERE the cache is warm on iteration 932 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-933] UNLESS the input is empty on iteration 933 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-934] IF the engine sees event 934 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-935] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 935.
[AC-936] WHEN the bench harness runs iteration 936 THE SYSTEM SHALL complete within 5 ms.
[AC-937] WHILE the matcher is hot on iteration 937 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-938] WHERE the cache is warm on iteration 938 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-939] UNLESS the input is empty on iteration 939 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-940] IF the engine sees event 940 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-941] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 941.
[AC-942] WHEN the bench harness runs iteration 942 THE SYSTEM SHALL complete within 5 ms.
[AC-943] WHILE the matcher is hot on iteration 943 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-944] WHERE the cache is warm on iteration 944 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-945] UNLESS the input is empty on iteration 945 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-946] IF the engine sees event 946 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-947] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 947.
[AC-948] WHEN the bench harness runs iteration 948 THE SYSTEM SHALL complete within 5 ms.
[AC-949] WHILE the matcher is hot on iteration 949 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-950] WHERE the cache is warm on iteration 950 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-951] UNLESS the input is empty on iteration 951 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-952] IF the engine sees event 952 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-953] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 953.
[AC-954] WHEN the bench harness runs iteration 954 THE SYSTEM SHALL complete within 5 ms.
[AC-955] WHILE the matcher is hot on iteration 955 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-956] WHERE the cache is warm on iteration 956 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-957] UNLESS the input is empty on iteration 957 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-958] IF the engine sees event 958 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-959] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 959.
[AC-960] WHEN the bench harness runs iteration 960 THE SYSTEM SHALL complete within 5 ms.
[AC-961] WHILE the matcher is hot on iteration 961 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-962] WHERE the cache is warm on iteration 962 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-963] UNLESS the input is empty on iteration 963 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-964] IF the engine sees event 964 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-965] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 965.
[AC-966] WHEN the bench harness runs iteration 966 THE SYSTEM SHALL complete within 5 ms.
[AC-967] WHILE the matcher is hot on iteration 967 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-968] WHERE the cache is warm on iteration 968 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-969] UNLESS the input is empty on iteration 969 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-970] IF the engine sees event 970 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-971] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 971.
[AC-972] WHEN the bench harness runs iteration 972 THE SYSTEM SHALL complete within 5 ms.
[AC-973] WHILE the matcher is hot on iteration 973 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-974] WHERE the cache is warm on iteration 974 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-975] UNLESS the input is empty on iteration 975 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-976] IF the engine sees event 976 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-977] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 977.
[AC-978] WHEN the bench harness runs iteration 978 THE SYSTEM SHALL complete within 5 ms.
[AC-979] WHILE the matcher is hot on iteration 979 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-980] WHERE the cache is warm on iteration 980 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-981] UNLESS the input is empty on iteration 981 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-982] IF the engine sees event 982 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-983] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 983.
[AC-984] WHEN the bench harness runs iteration 984 THE SYSTEM SHALL complete within 5 ms.
[AC-985] WHILE the matcher is hot on iteration 985 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-986] WHERE the cache is warm on iteration 986 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-987] UNLESS the input is empty on iteration 987 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-988] IF the engine sees event 988 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-989] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 989.
[AC-990] WHEN the bench harness runs iteration 990 THE SYSTEM SHALL complete within 5 ms.
[AC-991] WHILE the matcher is hot on iteration 991 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-992] WHERE the cache is warm on iteration 992 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-993] UNLESS the input is empty on iteration 993 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-994] IF the engine sees event 994 THEN the matcher THE SYSTEM SHALL return within 200 ns.
[AC-995] THE SYSTEM SHALL sustain at least 1 M iterations per second on iteration 995.
[AC-996] WHEN the bench harness runs iteration 996 THE SYSTEM SHALL complete within 5 ms.
[AC-997] WHILE the matcher is hot on iteration 997 THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-998] WHERE the cache is warm on iteration 998 THE SYSTEM SHALL reuse the in-memory buffer.
[AC-999] UNLESS the input is empty on iteration 999 THE SYSTEM SHALL CONTINUE TO return a non-empty result.
[AC-1000] IF the engine sees event 1000 THEN the matcher THE SYSTEM SHALL return within 200 ns.

## Out of Scope

- 不依赖任何运行时配置
- 不引用外部 spec 文件
- 不包含图片 / 表格等非纯文本元素
