#!/usr/bin/env python3
"""
Rewrite all_features_test.go mocks to exactly match real handler SQL.
Reads all handler files, extracts SQL, and generates correct mock expectations.
"""
import re, os, textwrap

BACKEND = "/root/project/backend"

# ============================================================
# STEP 1: Extract ALL SQL from handler files
# ============================================================
handler_sql = {}  # function_name -> list of (sql_string, is_query)

for fname in sorted(os.listdir(f"{BACKEND}/handlers")):
    if "_test" in fname or not fname.endswith(".go"): continue
    path = f"{BACKEND}/handlers/{fname}"
    with open(path) as f: content = f.read()
    
    lines = content.split('\n')
    current_func = None
    
    for i, line in enumerate(lines):
        # Track function
        fm = re.search(r'func\s+\((\w+)\s+\*?(\w+)\)\s+(\w+)\(', line)
        if fm:
            current_func = f"{fm.group(2)}.{fm.group(3)}"
            handler_sql.setdefault(current_func, [])
            continue
        fm2 = re.search(r'^func\s+(\w+)\(', line)
        if fm2 and current_func is None:
            current_func = fm2.group(1)
            handler_sql.setdefault(current_func, [])
        
        if current_func and ('QueryRow' in line or 'Query(' in line or 'Exec(' in line):
            # Find the SQL string (backtick or double-quoted)
            # Check for backtick strings on this line or following lines
            combined = line
            j = i + 1
            # If line has opening backtick but no closing, collect until closing
            if '`' in line and line.count('`') % 2 == 1:
                while j < len(lines) and '`' not in lines[j]:
                    combined += '\n' + lines[j]
                    j += 1
                if j < len(lines):
                    combined += '\n' + lines[j]
            
            bt_match = re.search(r'`([^`]+)`', combined)
            dq_matches = re.finditer(r'"([^"]*(?:SELECT|INSERT|UPDATE|DELETE)[^"]*)"', combined)
            
            if bt_match:
                sql_text = bt_match.group(1).strip()
                handler_sql[current_func].append((sql_text, 'QueryRow' in line or 'Query(' in line))
            else:
                for dq in dq_matches:
                    sql_text = dq.group(1).strip()
                    handler_sql[current_func].append((sql_text, 'QueryRow' in line or 'Query(' in line))

# ============================================================
# STEP 2: Build map of test function -> needed mocks
# ============================================================
# Group handler SQL by the test function that needs it
test_map = {
    "TestAuth_Register_Valid": ["AuthHandler.Register"],
    "TestAuth_Login_Valid": ["AuthHandler.Login"],
    "TestAuth_Login_WrongPassword": ["AuthHandler.Login"],
    "TestAuth_GetMe_Authorized": ["AuthHandler.GetProfile"],
    "TestAuth_UpdateProfile_Valid": ["AuthHandler.UpdateProfile"],
    "TestAuth_Logout_Valid": ["AuthHandler.Logout"],
    "TestProducts_GetCategories": ["GetCategories"],
    "TestProducts_GetProducts": ["GetProducts"],
    "TestProducts_GetProductBySlug": ["GetProduct"],
    "TestProducts_SearchProducts": ["GetProducts"],
    "TestAuth_CreateProduct_Unauthorized": ["CreateProduct"],
    "TestOrders_CreateOrder_Unauthorized": ["CreateOrder"],
    "TestOrders_GetUserOrders_Unauthorized": ["GetUserOrders"],
    "TestOrders_GetOrder_Unauthorized": ["GetOrder"],
    "TestCommunity_Profile": ["GetPublicProfile"],
    "TestCommunity_Feed": ["GetFeed"],
    "TestCommunity_CreatePost_Unauthorized": ["CreatePost"],
    "TestAdmin_BulkProductStatus": ["BulkProductHandler.AdminBulkUpdateProducts"],
    "TestAdmin_OrderStatusPaid": ["OrderStatusHandler.SetOrderStatus"],
    "TestAdmin_OrderStatusCompleted": ["OrderStatusHandler.SetOrderStatus"],
    "TestAdmin_OrderStatusCancelled": ["OrderStatusHandler.SetOrderStatus"],
    "TestAdmin_OrderStatusRefunded": ["OrderStatusHandler.SetOrderStatus"],
    "TestAdmin_BulkProductCategory": ["BulkProductHandler.AdminBulkUpdateCategory"],
    "TestAdmin_BulkProductToggle": ["BulkProductHandler.AdminBulkUpdateStatus"],
    "TestAdmin_DeleteUser": ["AuthHandler.DeleteUser"],
    "TestAdmin_ListUsers": ["AuthHandler.ListUsers"],
    "TestAdmin_DeleteCommunityPost": ["CommunityModerationHandler.DeletePost"],
    "TestAdmin_CommunityPostList": ["CommunityModerationHandler.ListCommunityPosts"],
    "TestAdmin_CommunityUserList": ["CommunityModerationHandler.ListCommunityUsers"],
    "TestAdmin_GuestOrderCheck": ["OrderStatusHandler.GuestOrderCheckSelf"],
    "TestAdmin_ExchangeRateList": ["ListExchangeRates"],
    "TestAdmin_ExchangeRateSet": ["SetExchangeRate"],
    "TestAdmin_Setting_Set": ["SetSetting"],
    "TestAdmin_Setting_Get": ["GetSetting"],
    "TestAdmin_Stats": ["GetProductStats"],
    "TestCart_GetOrCreate": ["CartHandler.GetOrCreateCart"],
    "TestCart_AddItem": ["CartItemHandler.AddToCart"],
    "TestCart_RemoveItem": ["CartItemHandler.RemoveFromCart"],
    "TestWishlist_Toggle": ["WishlistHandler.ToggleWishlist"],
    "TestWishlist_List": ["WishlistHandler.WishlistList"],
    "TestReviews_Create": ["ReviewHandler.CreateReview"],
    "TestReviews_List": ["ReviewHandler.ReviewList"],
    "TestRecentlyViewed_Record": ["RecentlyViewedHandler.RecordView"],
    "TestRecentlyViewed_List": ["RecentlyViewedHandler.RecentlyViewedList"],
    "TestComparison_Toggle": ["ComparisonHandler.ToggleComparison"],
    "TestComparison_List": ["ComparisonHandler.ComparisonList"],
    "TestRecommendations_Get": ["RecommendationHandler.Recommend"],
    "TestGuestOrder_Create": ["CreateGuestOrder"],
    "TestGuestOrder_Check": ["GetGuestOrder"],
    "TestOrderStatusCheck": ["CheckPaymentStatus"],
    "TestProducts_CreateProduct": ["CreateProduct"],
    "TestSettings_List": ["ListSettings"],
    "TestSettings_Get": ["GetSetting"],
    "TestExchangeRates_GetAll": ["ListExchangeRates"],
    "TestExchangeRates_GetOne": ["GetExchangeRate"],
    "TestAdminOrders_OrderPayment": ["AdminOrderDetail"],
    "TestAdminOrders_ListAllOrders": ["ListAllOrders"],
    "TestNewProducts_List": ["BulkProductHandler.AdminProductList"],
    "TestNewProducts_Delete": ["DeleteProduct"],
    "TestDBSchema_Migration011Tables": ["GetHealth"],
}

# ============================================================
# STEP 3: Generate correct mock SQL for each test function
# ============================================================
def sql_to_mock_pattern(sql):
    """Convert a SQL string to a sqlmock regex pattern."""
    # Escape backticks and special chars for Go raw string
    # sqlmock uses regex, so we need to escape regex specials
    # But $1, $2 are already regex-friendly
    s = sql.replace('`', '')
    # The pattern in the test is inside a Go raw string (backticks)
    # So backslashes in the Go source are literal
    # For regex: \$ matches literal $, \( matches literal (, etc.
    # In Go raw string: \$ is two chars: backslash + dollar
    # sqlmock receives the string as-is and compiles as regex
    # So we need: \$ in Go source -> \$ in regex -> matches literal $
    # Currently broken: \\$ in Go source -> \\ in regex -> matches literal backslash + $ (wrong!)
    # Fix: replace \\$ with \$ throughout
    s = s.replace('\\\\$', '$')  # Fix double-escaped dollar
    s = s.replace('\\\\*', '*')  # Fix double-escaped star
    return s

def sql_to_columns(sql):
    """Extract column names from SELECT statement."""
    s = sql.strip()
    # Handle multi-line SELECT
    s = re.sub(r'\s+', ' ', s)
    m = re.search(r'SELECT\s+(.*?)\s+FROM', s, re.IGNORECASE | re.DOTALL)
    if not m:
        return []
    
    cols_str = m.group(1)
    # Parse columns: handle "expr as alias", "*", function calls
    columns = []
    for part in cols_str.split(','):
        part = part.strip()
        if not part: continue
        # Check for "expr as alias"
        am = re.search(r'(.+?)\s+as\s+(\w+)', part, re.IGNORECASE)
        if am:
            columns.append(am.group(2))
        elif part == '*':
            columns.append("*")
        else:
            # Take last identifier
            tokens = part.split()
            columns.append(tokens[-1])
    return columns

def generate_mock_expectation(func_name, sql, is_query, row_data=None):
    """Generate a Go mock expectation line."""
    # Clean SQL for Go raw string
    clean_sql = sql.strip()
    # Remove newlines and extra whitespace for the regex pattern
    pattern = re.sub(r'\s+', ' ', clean_sql)
    
    prefix = "mock.ExpectQuery" if is_query else "mock.ExpectExec"
    
    # Determine args needed
    param_count = len(re.findall(r'\$(\d+)', pattern))
    
    lines = []
    lines.append(f"\t\t{prefix}(`{pattern}`).")
    
    if row_data:
        if is_query:
            cols = sql_to_columns(sql)
            if not cols:
                cols = [f"col{i}" for i in range(1, pattern.count('$')+1)]
            lines.append(f"\t\t\tWillReturnRows(sqlmock.NewRows([]string{{{','.join(repr(c) for c in cols)}}}).")
            if isinstance(row_data, list):
                for row in row_data:
                    vals = ", ".join(repr(v) if v is not None else "nil" for v in row)
                    lines.append(f"\t\t\t\tAddRow({vals})).")
            else:
                lines.append(f"\t\t\t\tAddRow({row_data}).")
        else:
            lines.append(f"\t\t\tWillReturnResult(sqlmock.NewResult({row_data[0]}, {row_data[1])}).")
    
    # Remove trailing ")." and replace with just ")"
    if lines[-1].endswith(")."):
        lines[-1] = lines[-1][:-2] + ")"
    
    return "\n".join(lines)

# ============================================================
# STEP 4: Generate the fixed test file
# ============================================================
print("Generating fixed test file...")

# Read current test file to preserve non-mock parts
test_path = f"{BACKEND}/handlers/all_features_test.go"
with open(test_path) as f:
    current_test = f.read()

# For now, let's just fix the most critical failing tests by patching specific mocks
# We'll generate the mock patches programmatically

fixes = []

for test_func, handler_funcs in test_map.items():
    sql_list = []
    for hf in handler_funcs:
        if hf in handler_sql:
            sql_list.extend(handler_sql[hf])
    
    if not sql_list:
        print(f"  SKIP {test_func}: no SQL found for {handler_funcs}")
        continue
    
    # Generate mock expectations
    mock_code = []
    for sql, is_query in sql_list:
        mock_code.append(generate_mock_expectation(test_func, sql, is_query))
    
    if mock_code:
        fixes.append((test_func, mock_code))

print(f"\nGenerated mock fixes for {len(fixes)} test functions")
print("\nSample fix for TestProducts_GetProducts:")
if "TestProducts_GetProducts" in [f[0] for f in fixes]:
    for mc in [f[1] for f in fixes if f[0] == "TestProducts_GetProducts"][0]:
        print(mc[:150])
