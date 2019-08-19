# go-odata

A golang package to create ODATA REST APIs using mongodb

## commands and Examples

- Select: selects specified columns for the data records. This uses a string split with a comma delimiter.
EX. http://localhost/test?$select=name,age

- Top: returns the top x records where x is a valid integer value
EX: http://localhost/test?$top=10

- Skip: skips the first y records where y is a valid integer value
EX: http://localhost/test?$skip=5

- Count: returns an integer value that equals the total count of records found in the collection. This command requires no parameters and takes precedence over all other commands.
EX: http://localhost/test?$count

- OrderBy: returns the collection in an order based on the input parameters. The input is a string with comma delimiters and uses a similar string parsing method as select. The order by parameter also supports ascending (asc) and descending (desc) options as part of each column parameter.
EX: http://localhost/test?$orderby=name asc, age desc

- InlineCount: returns the query result records along with the count. The inlinecount parameter takes either 'allpages' or 'none' as the input. Any other input will cause the count to not return.
EX: http://localhost/test?$skip=5&$inlinecount=allpages

- Filter: Returns data based on the expression input by the user. The parser utilizes its own library to define keywords and regular expressions to sort the input. The input is then put into a tree structure which can be converted into a map of interfaces. The map structure allows the database adapters to translate the input into the appropriate queries.
EX: http://localhost/test?$filter=name eq 'val'

See ODATA specification [https://www.odata.org/](https://www.odata.org/)