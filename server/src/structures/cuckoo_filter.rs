use rand::Rng;
use std::collections::hash_map::DefaultHasher;
use std::hash::{Hash, Hasher};

#[derive(Clone)]
struct Bucket {
    fingerprints: Vec<Option<u64>>,
    size: usize,
}

impl Bucket {
    fn new(size: usize) -> Self {
        Bucket {
            fingerprints: vec![None; size],
            size,
        }
    }

    fn insert(&mut self, fingerprint: u64) -> bool {
        for fp in &mut self.fingerprints {
            if fp.is_none() {
                *fp = Some(fingerprint);
                return true;
            }
        }
        false
    }

    fn remove(&mut self, fingerprint: u64) -> bool {
        for fp in &mut self.fingerprints {
            if *fp == Some(fingerprint) {
                *fp = None;
                return true;
            }
        }
        false
    }

    fn contains(&self, fingerprint: u64) -> bool {
        self.fingerprints.contains(&Some(fingerprint))
    }
}

pub struct CuckooFilter {
    buckets: Vec<Bucket>,
    bucket_size: usize,
    max_kicks: usize,
}

impl CuckooFilter {
    pub fn new(num_buckets: usize, bucket_size: usize, max_kicks: usize) -> Self {
        let buckets = vec![Bucket::new(bucket_size); num_buckets];
        CuckooFilter {
            buckets,
            bucket_size,
            max_kicks,
        }
    }

    fn hash<T: Hash>(&self, item: &T) -> u64 {
        let mut hasher = DefaultHasher::new();
        item.hash(&mut hasher);
        hasher.finish()
    }

    fn fingerprint(&self, hash: u64) -> u64 {
        hash & 0xFFFF_FFFF_FFFF_FFFF
    }

    fn index(&self, hash: u64) -> usize {
        (hash as usize) % self.buckets.len()
    }

    fn alt_index(&self, index: usize, fingerprint: u64) -> usize {
        (index ^ (self.hash(&fingerprint) as usize)) % self.buckets.len()
    }

    pub fn add<T: Hash>(&mut self, item: &T) -> bool {
        let hash = self.hash(item);
        let fingerprint = self.fingerprint(hash);
        let mut i1 = self.index(hash);
        let mut i2 = self.alt_index(i1, fingerprint);

        if self.buckets[i1].insert(fingerprint) || self.buckets[i2].insert(fingerprint) {
            return true;
        }

        let mut rng = rand::thread_rng();
        for _ in 0..self.max_kicks {
            let index = if rng.gen_bool(0.5) { i1 } else { i2 };
            let bucket = &mut self.buckets[index];
            let pos = rng.gen_range(0..self.bucket_size);
            let evicted_fp = bucket.fingerprints[pos].take().unwrap();
            bucket.fingerprints[pos] = Some(fingerprint);

            i1 = index;
            i2 = self.alt_index(i1, evicted_fp);
            if self.buckets[i2].insert(evicted_fp) {
                return true;
            }
        }
        false
    }

    pub fn contains<T: Hash>(&self, item: &T) -> bool {
        let hash = self.hash(item);
        let fingerprint = self.fingerprint(hash);
        let i1 = self.index(hash);
        let i2 = self.alt_index(i1, fingerprint);

        self.buckets[i1].contains(fingerprint) || self.buckets[i2].contains(fingerprint)
    }

    pub fn remove<T: Hash>(&mut self, item: &T) -> bool {
        let hash = self.hash(item);
        let fingerprint = self.fingerprint(hash);
        let i1 = self.index(hash);
        let i2 = self.alt_index(i1, fingerprint);

        self.buckets[i1].remove(fingerprint) || self.buckets[i2].remove(fingerprint)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_cuckoo_filter_add_contains() {
        let mut filter = CuckooFilter::new(100, 4, 500);

        assert!(filter.add(&"hello"), "Failed to add 'hello'");
        assert!(filter.add(&"world"), "Failed to add 'world'");

        assert!(filter.contains(&"hello"), "'hello' should be in the filter");
        assert!(filter.contains(&"world"), "'world' should be in the filter");
        assert!(!filter.contains(&"rust"), "'rust' should not be in the filter");
    }

    #[test]
    fn test_cuckoo_filter_remove() {
        let mut filter = CuckooFilter::new(100, 4, 500);

        filter.add(&"example");
        assert!(filter.contains(&"example"), "'example' should be in the filter");

        assert!(filter.remove(&"example"), "Failed to remove 'example'");
        assert!(!filter.contains(&"example"), "'example' should no longer be in the filter");
    }

    #[test]
    fn test_cuckoo_filter_collision() {
        let mut filter = CuckooFilter::new(10, 2, 10);

        filter.add(&"one");
        filter.add(&"two");

        // Test for possible collisions
        assert!(!filter.contains(&"three"), "'three' should not be in the filter");
    }
}
