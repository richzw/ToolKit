package codes

class Solution {
    public boolean isMatch(String s, String p) {
        if (s.equals(p)) {
            return true;
        }

        boolean isFirstMatch = false;
        if (!s.isEmpty() && !p.isEmpty() && (s.charAt(0) == p.charAt(0) || p.charAt(0) == '.')) {
            isFirstMatch = true;
        }

        if (p.length() >= 2 && p.charAt(1) == '*') {
// 看 s[i,...n] 和 p[j+2,...m] 或者是 s[i+1,...n] 和 p[j,...m]
            return isMatch(s, p.substring(2))
                    || (isFirstMatch && isMatch(s.substring(1), p));
        }

// 看 s[i+1,...n] 和 p[j+1,...m]
        return isFirstMatch && isMatch(s.substring(1), p.substring(1));
    }

    public boolean isMatch(String s, String p) {
        if (s.equals(p)) {
            return true;
        }

        boolean[] memo = new boolean[s.length() + 1];

        return helper(s.toCharArray(), p.toCharArray(),
                s.length() - 1, p.length() - 1, memo);
    }

    private boolean helper(char[] s, char[] p, int i, int j, boolean[] memo) {
        if (memo[i + 1]) {
            return true;
        }

        if (i == -1 && j == -1) {
            memo[i + 1] = true;
            return true;
        }

        boolean isFirstMatching = false;

        if (i >= 0 && j >= 0 && (s[i] == p[j] || p[j] == '.'
                || (p[j] == '*' && (p[j - 1] == s[i] || p[j - 1] == '.')))) {
            isFirstMatching = true;
        }

        if (j >= 1 && p[j] == '*') {
// 看 s[0,...i] 和 p[0,...j-2]
            boolean zero = helper(s, p, i, j - 2, memo);
// 看 s[0,...i-1] 和 p[0,...j]
            boolean match = isFirstMatching && helper(s, p, i - 1, j, memo);

            if (zero || match) {
                memo[i + 1] = true;
            }

            return memo[i + 1];
        }

// 看 s[0,...i-1] 和 p[0,...j-1]
        if (isFirstMatching && helper(s, p, i - 1, j - 1, memo)) {
            memo[i + 1] = true;
        }

        return memo[i + 1];
    }

    public boolean isMatch(String s, String p) {
        if (s.equals(p)) {
            return true;
        }

        char[] sArr = s.toCharArray();
        char[] pArr = p.toCharArray();

// dp[i][j] => is s[0, i - 1] match p[0, j - 1] ?
        boolean[][] dp = new boolean[sArr.length + 1][pArr.length + 1];

        dp[0][0] = true;

        for (int i = 1; i <= pArr.length; ++i) {
            dp[0][i] = pArr[i - 1] == '*' ? dp[0][i - 2] : false;
        }

        for (int i = 1; i <= sArr.length; ++i) {
            for (int j = 1; j <= pArr.length; ++j) {
                if (sArr[i - 1] == pArr[j - 1] || pArr[j - 1] == '.') {
// 看 s[0,...i-1] 和 p[0,...j-1]
                    dp[i][j] = dp[i - 1][j - 1];
                }

                if (pArr[j - 1] == '*') {
// 看 s[0,...i] 和 p[0,...j-2]
                    dp[i][j] |= dp[i][j - 2];

                    if (pArr[j - 2] == sArr[i - 1] || pArr[j - 2] == '.') {
// 看 s[0,...i-1] 和 p[0,...j]
                        dp[i][j] |= dp[i - 1][j];
                    }
                }
            }
        }

        return dp[sArr.length][pArr.length];
    }
}