package render

import "strings"

// sqlKeywords are highlighted in SQL code blocks. Types are included so
// a CREATE TABLE reads as one colored shape.
var sqlKeywords = map[string]bool{}

func init() {
	for _, w := range strings.Fields(`
		SELECT FROM WHERE AND OR NOT NULL IS IN LIKE ILIKE BETWEEN EXISTS AS ON
		JOIN LEFT RIGHT INNER OUTER FULL CROSS USING GROUP BY ORDER HAVING LIMIT
		OFFSET DESC ASC DISTINCT UNION ALL ANY SOME CASE WHEN THEN ELSE END
		INSERT INTO VALUES UPDATE SET DELETE CREATE TABLE INDEX UNIQUE DROP ALTER
		ADD COLUMN CONSTRAINT PRIMARY KEY FOREIGN REFERENCES DEFAULT CHECK
		GENERATED ALWAYS IDENTITY IF EXPLAIN ANALYZE VERBOSE VACUUM FULL
		BEGIN COMMIT ROLLBACK TRANSACTION ISOLATION LEVEL READ COMMITTED
		REPEATABLE SERIALIZABLE FOR SHARE NOWAIT SKIP LOCKED RETURNING WITH
		RECURSIVE OVER PARTITION WINDOW ROWS RANGE PRECEDING FOLLOWING CURRENT
		ROW CAST TRUE FALSE GRANT REVOKE POLICY ENABLE SECURITY TO CONCURRENTLY
		CASCADE ARRAY INTERVAL LATERAL NULLS FIRST LAST ONLY VIEW MATERIALIZED
		FUNCTION RETURNS LANGUAGE TRIGGER BEFORE AFTER EACH EXECUTE PROCEDURE
		BIGINT INT INTEGER SMALLINT SERIAL BIGSERIAL TEXT VARCHAR CHAR BOOLEAN
		BOOL NUMERIC DECIMAL REAL FLOAT DOUBLE PRECISION TIMESTAMP TIMESTAMPTZ
		DATE TIME JSON JSONB UUID BYTEA`) {
		sqlKeywords[w] = true
	}
}

// highlightSQL colors one line of SQL: keywords bold magenta, strings
// green, numbers yellow, comments dim, function calls cyan, psql
// meta-commands yellow. Identifiers stay plain so they stand apart.
func highlightSQL(line string) string {
	if strings.HasPrefix(strings.TrimLeft(line, " "), `\`) {
		return ansiYellow + line + ansiReset
	}
	var b strings.Builder
	n := len(line)
	for i := 0; i < n; {
		c := line[i]
		switch {
		case c == '-' && i+1 < n && line[i+1] == '-':
			b.WriteString(ansiDim + line[i:] + ansiReset)
			i = n
		case c == '\'':
			j := i + 1
			for j < n {
				if line[j] == '\'' {
					if j+1 < n && line[j+1] == '\'' {
						j += 2 // escaped quote inside the string
						continue
					}
					break
				}
				j++
			}
			if j < n {
				j++ // closing quote
			}
			b.WriteString(ansiGreen + line[i:j] + ansiReset)
			i = j
		case isIdentStart(c):
			j := i
			for j < n && isIdentChar(line[j]) {
				j++
			}
			word := line[i:j]
			switch {
			case sqlKeywords[strings.ToUpper(word)]:
				b.WriteString(ansiBold + ansiMagenta + word + ansiReset)
			case j < n && line[j] == '(':
				b.WriteString(ansiCyan + word + ansiReset)
			default:
				b.WriteString(word)
			}
			i = j
		case c >= '0' && c <= '9':
			j := i
			for j < n && (line[j] >= '0' && line[j] <= '9' || line[j] == '.') {
				j++
			}
			b.WriteString(ansiYellow + line[i:j] + ansiReset)
			i = j
		default:
			b.WriteByte(c)
			i++
		}
	}
	return b.String()
}

func isIdentStart(c byte) bool {
	return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isIdentChar(c byte) bool {
	return isIdentStart(c) || c >= '0' && c <= '9'
}
