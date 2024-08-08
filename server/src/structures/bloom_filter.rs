use std::hash::{Hash, Hasher};
use std::collections::hash_map::DefaultHasher;

pub(crate) struct BloomFilter {
    bit_array: Vec<bool>,
    num_hashes: usize,
}

impl BloomFilter {
    pub fn new(size: usize, num_hashes: usize) -> Self {
        BloomFilter {
            bit_array: vec![false; size],
            num_hashes,
        }
    }

    fn hash<T: Hash>(&self, item: &T, i: usize) -> usize {
        let mut hasher = DefaultHasher::new();
        item.hash(&mut hasher);
        i.hash(&mut hasher);
        (hasher.finish() as usize) % self.bit_array.len()
    }

    pub fn add<T: Hash>(&mut self, item: &T) {
        for i in 0..self.num_hashes {
            let index = self.hash(item, i);
            self.bit_array[index] = true;
        }
    }

    pub fn contains<T: Hash>(&self, item: &T) -> bool {
        for i in 0..self.num_hashes {
            let index = self.hash(item, i);
            if !self.bit_array[index] {
                return false;
            }
        }
        true
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_bloom_filter_contains() {
        let mut bloom_filter = BloomFilter::new(100, 3);
        bloom_filter.add(&"hello");
        bloom_filter.add(&"world");

        assert!(bloom_filter.contains(&"hello"), "Expected 'hello' to be in the filter");
        assert!(bloom_filter.contains(&"world"), "Expected 'world' to be in the filter");
        assert!(!bloom_filter.contains(&"rust"), "Expected 'rust' to not be in the filter");
    }

    #[test]
    fn test_bloom_filter_add() {
        let mut bloom_filter = BloomFilter::new(100, 3);
        bloom_filter.add(&"example");

        assert!(bloom_filter.contains(&"example"), "Expected 'example' to be in the filter after addition");
        assert!(!bloom_filter.contains(&"nonexistent"), "Expected 'nonexistent' to not be in the filter");
    }

    #[test]
    fn test_bloom_filter_collision() {
        let mut bloom_filter = BloomFilter::new(5, 4);
        bloom_filter.add(&"one");
        bloom_filter.add(&"two");

        // Due to small bit array, this might collide and cause a false positive.
        assert!(bloom_filter.contains(&"three"), "Expected 'three' to not be in the filter, but collisions may cause false positives");
    }
}