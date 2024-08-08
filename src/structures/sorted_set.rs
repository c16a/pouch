use std::collections::{BTreeMap, HashMap};
use std::path::Component::ParentDir;

pub(crate) struct SortedSet<T> {
    score_to_elements: BTreeMap<i64, Vec<T>>,
    element_to_score: HashMap<T, i64>,
}

#[derive(Copy, Clone)]
pub(crate) enum SortedSetAddReturnType {
    Added,   // ZAdd would only return the number of new elements added
    Changed, // ZAdd would return sum of new elements added and existing elements with updated scores
}

impl<T: Ord + Clone + Eq + std::hash::Hash> SortedSet<T> {
    pub fn new() -> Self {
        SortedSet {
            score_to_elements: BTreeMap::new(),
            element_to_score: HashMap::new(),
        }
    }

    pub fn add_all(
        &mut self,
        values: HashMap<T, i64>,
        return_type: Option<SortedSetAddReturnType>,
    ) -> i64 {
        let inserted_rows = values.into_iter().fold(0, |acc, value| {
            acc + self.add(value.0, value.1, return_type)
        });
        inserted_rows
    }

    pub fn add(
        &mut self,
        element: T,
        score: i64,
        return_type: Option<SortedSetAddReturnType>,
    ) -> i64 {
        let return_type = return_type.unwrap_or_else(|| SortedSetAddReturnType::Added);

        let mut affected_rows: i64 = 1;

        if let Some(&old_score) = self.element_to_score.get(&element) {
            let score_changed = old_score != score;

            match return_type {
                SortedSetAddReturnType::Added => {
                    affected_rows = 0;
                }
                SortedSetAddReturnType::Changed => {
                    affected_rows = if score_changed { 1 } else { 0 };
                }
            }

            self.remove(&element);
        }

        self.score_to_elements
            .entry(score)
            .or_insert(Vec::new())
            .push(element.clone());
        self.element_to_score.insert(element, score);

        affected_rows
    }

    pub fn remove(&mut self, element: &T) -> Option<i64> {
        if let Some(score) = self.element_to_score.remove(element) {
            if let Some(elements) = self.score_to_elements.get_mut(&score) {
                elements.retain(|e| e != element);
                if elements.is_empty() {
                    self.score_to_elements.remove(&score);
                }
            }
            return Some(score);
        }
        None
    }

    pub fn score(&self, element: &T) -> Option<i64> {
        self.element_to_score.get(element).copied()
    }

    pub fn elements(&self) -> Vec<&T> {
        self.score_to_elements
            .values()
            .flat_map(|elements| elements.iter())
            .collect()
    }

    pub fn cardinality(&self) -> usize {
        self.element_to_score.len()
    }

    pub fn range(&self, min: i64, max: i64) -> Vec<&T> {
        self.score_to_elements
            .range(min..=max)
            .flat_map(|(_, elements)| elements.iter())
            .collect()
    }
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn test_sorted_set() {
        let mut sorted_set = SortedSet::new();

        let mut affected_rows = sorted_set.add("apple", 10, Some(SortedSetAddReturnType::Added));
        assert_eq!(affected_rows, 1);

        affected_rows = sorted_set.add("banana", 20, Some(SortedSetAddReturnType::Added));
        assert_eq!(affected_rows, 1);

        affected_rows = sorted_set.add("apple", 15, Some(SortedSetAddReturnType::Added));
        assert_eq!(affected_rows, 0);

        affected_rows = sorted_set.add("apple", 25, Some(SortedSetAddReturnType::Changed));
        assert_eq!(affected_rows, 1);
    }
}
