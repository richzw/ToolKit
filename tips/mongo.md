
- [Index](https://www.mongodb.com/blog/post/performance-best-practices-indexing)
  - Use Compound Indexes
    
    For compound indexes, follow [ESR rule](https://www.alexbevi.com/blog/2020/05/16/optimizing-mongodb-compound-indexes-the-equality-sort-range-esr-rule/)
    - First, add those fields against which **Equality** queries are run.
    - The next fields to be indexed should reflect the **Sort** order of the query.
    - The last fields represent the **Range** of data to be accessed.
    
    Operator Type Check 
    - Inequality - _$ne $nin_ belong to **Range**
    - Regex - _/car/_ belong to **Range**
    - _$in_ 
      - Alone: a series of **Equality** matches
      - Combined: possible a **Range**
        - May optimize **blocking sort** with **Merge Sort**
    
    Exception
      - check **totalDocsExamined**
  - Use Covered Queries When Possible
    - Covered queries return results from an index directly without having to access the source documents, and are therefore very efficient.
    - If the _explain()_ output displays **totalDocsExamined** as _0_, this shows the query is covered by an index.
    - A common gotcha when trying to achieve covered queries is that the __id_ field is always returned by default. You need to explicitly exclude it from query results, or add it to the index.
  - Use Caution When Considering Indexes on Low-Cardinality Fields
    - Queries on fields with a small number of unique values (low cardinality) can return large result sets. Compound indexes may include fields with low cardinality, but the value of the combined fields should exhibit high cardinality.
  - Wildcard Indexes Are Not a Replacement for Workload-Based Index Planning
    - If your application’s query patterns are known in advance, then you should use more selective indexes on the specific fields accessed by the queries
  - Use text search to match words inside a field
    - If you only want to match on a specific word in a field with a lot of text, then use a text index.
    - If you are running MongoDB in the Atlas service, consider using Atlas Full Text Search which provides a fully-managed Lucene index integrated with the MongoDB database.
  - Use Partial Indexes
    - Reduce the size and performance overhead of indexes by only including documents that will be accessed through the index
  - Take Advantage of Multi-Key Indexes for Querying Arrays
    - If your query patterns require accessing individual array elements, use a multi-key index.
  - Avoid Regular Expressions That Are Not Left Anchored or Rooted
    - Indexes are ordered by value. Leading wildcards are inefficient and may result in full index scans. Trailing wildcards can be efficient if there are sufficient case-sensitive leading characters in the expression.
  - Avoid Case Insensitive Regular Expressions
    - If the sole reason for using a regex is case insensitivity, use a case insensitive index instead, as those are faster.
  - Use Index Optimizations Available in the WiredTiger Storage Engine
    - If you are self-managing MongoDB, you can optionally place indexes on their own separate volume, allowing for faster disk paging and lower contention. See wiredTiger options for more information.

