A Bloom filter is a space-efficient probabilistic data structure used to determine whether an element belongs to a set or not. The Bloom filter allows false positives (maybe in the set) but never false negatives (definitely not in set) . If you are new to bloomfilters, give [Bloom Filters by Example](https://llimllib.github.io/bloomfilter-tutorial/) a read.

The `bloomfilter` package is suitable for caching filtering, decentralized aggregation, search large chemical structure databases and many other applications. More specifically, we use this package in production with KrakenD to [distributedly reject JWT tokens](https://www.krakend.io/docs/authorization/revoking-tokens/) as it allows us to perform massive rejections with very little memory consumption. For instance, 100 million tokens of any size consume around 0.5GB RAM (with a rate of false positives of 1 in 999,925,224 tokens), and lookups are completed in constant time (k number of hashes). These numbers are impossible to get with a key-value or a relational database.

## Implementations
This repository contains several bloomfilter implementations that you can use to solve different distributed computing problems. The solution starts from an optimized local implementation that adds rotation, RPC coordination, and generic rejecters. The packages are:

- `bitset`: Implementations of bitsets for basic sets.
- `bloomfilter`: Optimized implementation of the bloomfilter.
- `rotable`: Implementation over the BF with 3 rotating buckets.
- `rpc`: Implementation of an RPC layer over rotable.
- `krakend`: Integration of the `rpc` package as a rejecter for KrakenD

