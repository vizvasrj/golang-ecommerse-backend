GET    /api/ping                 --> main.main.func1 (5 handlers)
POST   /api/auth/login           --> src/pkg/module/auth.SetupRouter.Login.func1 (5 handlers)
POST   /api/auth/register        --> src/pkg/module/auth.SetupRouter.Register.func2 (5 handlers)
GET    /api/auth/google          --> src/pkg/module/auth.SetupRouter.GoogleLogin.func3 (5 handlers)
POST   /api/auth/google          --> src/pkg/module/auth.SetupRouter.GoogleCallbackPOST.func4 (5 handlers)
GET    /api/auth/google/callback --> src/pkg/module/auth.SetupRouter.GoogleCallback.func5 (5 handlers)
POST   /api/auth/forgot          --> src/pkg/module/auth.SetupRouter.ForgotPassword.func6 (5 handlers)
POST   /api/auth/reset/:token    --> src/pkg/module/auth.SetupRouter.ResetPasswordFromToken.func7 (5 handlers)
POST   /api/auth/reset           --> src/pkg/module/auth.SetupRouter.ResetPassword.func9 (6 handlers)
POST   /api/address/add          --> src/pkg/module/address.SetupRouter.AddAddress.func2 (6 handlers)
GET    /api/address              --> src/pkg/module/address.SetupRouter.GetAddresses.func3 (6 handlers)
GET    /api/address/:id          --> src/pkg/module/address.SetupRouter.GetAddress.func4 (6 handlers)
PUT    /api/address/:id          --> src/pkg/module/address.SetupRouter.UpdateAddress.func5 (6 handlers)
DELETE /api/address/delete/:id   --> src/pkg/module/address.SetupRouter.DeleteAddress.func6 (6 handlers)
PUT    /api/address/default/:id  --> src/pkg/module/address.SetupRouter.SetDefaultAddress.func7 (6 handlers)
POST   /api/brand/add            --> src/pkg/module/brand.SetupRouter.AddBrand.func3 (7 handlers)
GET    /api/brand/list           --> src/pkg/module/brand.SetupRouter.ListBrands.func4 (5 handlers)
GET    /api/brand                --> src/pkg/module/brand.SetupRouter.ListBrands.func5 (5 handlers)
GET    /api/brand/:id            --> src/pkg/module/brand.SetupRouter.GetBrandByID.func6 (5 handlers)
GET    /api/brand/list/select    --> src/pkg/module/brand.SetupRouter.ListSelectBrands.func7 (5 handlers)
PUT    /api/brand/:id            --> src/pkg/module/brand.SetupRouter.UpdateBrand.func10 (7 handlers)
PUT    /api/brand/:id/active     --> src/pkg/module/brand.SetupRouter.UpdateBrandActive.func13 (7 handlers)
DELETE /api/brand/delete/:id     --> src/pkg/module/brand.SetupRouter.DeleteBrand.func16 (7 handlers)
GET    /api/product/item/:slug   --> src/pkg/module/product.SetupRouter.GetProductBySlug.func1 (5 handlers)

✅ GET    /api/product/list/search/:name --> src/pkg/module/product.SetupRouter.SearchProductsByName.func2 (5 handlers)
GET    /api/product/list         --> src/pkg/module/product.SetupRouter.FetchStoreProductsByFilters.func3 (5 handlers)
GET    /api/product/list/select  --> src/pkg/module/product.SetupRouter.FetchProductNames.func4 (5 handlers)
POST   /api/product/add          --> src/pkg/module/product.SetupRouter.AddProduct.func7 (7 handlers)
GET    /api/product              --> src/pkg/module/product.SetupRouter.FetchProducts.func10 (7 handlers)
GET    /api/product/:id          --> src/pkg/module/product.SetupRouter.FetchProduct.func13 (7 handlers)
PUT    /api/product/:id          --> src/pkg/module/product.SetupRouter.UpdateProduct.func16 (7 handlers)
PUT    /api/product/:id/active   --> src/pkg/module/product.SetupRouter.UpdateProductStatus.func19 (7 handlers)
DELETE /api/product/delete/:id   --> src/pkg/module/product.SetupRouter.DeleteProduct.func22 (7 handlers)
GET    /api/user/search          --> src/pkg/module/user.SetupRouter.SearchUsers.func3 (7 handlers)
GET    /api/user                 --> src/pkg/module/user.SetupRouter.FetchUsers.func6 (7 handlers)
GET    /api/user/me              --> src/pkg/module/user.SetupRouter.GetCurrentUser.func8 (6 handlers)
PUT    /api/user                 --> src/pkg/module/user.SetupRouter.UpdateUserProfile.func10 (6 handlers)
POST   /api/merchant/add         --> src/pkg/module/merchant.SetupRouter.AddMerchant.func2 (6 handlers)
GET    /api/merchant/search      --> src/pkg/module/merchant.SetupRouter.SearchMerchants.func5 (7 handlers)
GET    /api/merchant             --> src/pkg/module/merchant.SetupRouter.FetchAllMerchants.func8 (7 handlers)
PUT    /api/merchant/:id/active  --> src/pkg/module/merchant.SetupRouter.DisableMerchantAccount.func11 (7 handlers)
PUT    /api/merchant/approve/:id --> src/pkg/module/merchant.SetupRouter.ApproveMerchant.func14 (7 handlers)
POST   /api/category/add         --> src/pkg/module/category.SetupRoute.AddCategory.func3 (7 handlers)
GET    /api/category/list        --> src/pkg/module/category.SetupRoute.ListCategories.func4 (5 handlers)
GET    /api/category             --> src/pkg/module/category.SetupRoute.FetchCategories.func5 (5 handlers)
GET    /api/category/:id         --> src/pkg/module/category.SetupRoute.FetchCategory.func6 (5 handlers)
PUT    /api/category/:id         --> src/pkg/module/category.SetupRoute.UpdateCategory.func9 (7 handlers)
PUT    /api/category/:id/active  --> src/pkg/module/category.SetupRoute.UpdateCategoryStatus.func12 (7 handlers)
DELETE /api/category/delete/:id  --> src/pkg/module/category.SetupRoute.DeleteCategory.func15 (7 handlers)
PUT    /api/category/product/:product_id/add --> src/pkg/module/category.SetupRoute.AddProductToCategory.func18 (7 handlers)
DELETE /api/category/:category_id/product/:product_id --> src/pkg/module/category.SetupRoute.RemoveProductFromCategory.func21 (7 handlers)
POST   /api/cart/add             --> src/pkg/module/cart.SetupRoute.AddToCart.func2 (6 handlers)
DELETE /api/cart/delete/:cartId  --> src/pkg/module/cart.SetupRoute.DeleteCart.func4 (6 handlers)
POST   /api/cart/add/:cartId     --> src/pkg/module/cart.SetupRoute.AddProductToCart.func6 (6 handlers)
POST   /api/cart/add_or_update   --> src/pkg/module/cart.SetupRoute.AddProductToCartV2.func8 (6 handlers)
DELETE /api/cart/delete/:cartId/:productId --> src/pkg/module/cart.SetupRoute.RemoveProductFromCart.func10 (6 handlers)
GET    /api/cart/:cartId         --> src/pkg/module/cart.SetupRoute.GetCartByCartID.func12 (6 handlers)
POST   /api/order/add            --> src/pkg/module/order.SetupRoute.AddOrderWithCartItemAndAddress.func2 (6 handlers)
GET    /api/order/search         --> src/pkg/module/order.SetupRoute.SearchOrders.func4 (6 handlers)
GET    /api/order                --> src/pkg/module/order.SetupRoute.FetchOrders.func6 (6 handlers)
GET    /api/order/me             --> src/pkg/module/order.SetupRoute.FetchUserOrders.func8 (6 handlers)
GET    /api/order/:orderId       --> src/pkg/module/order.SetupRoute.FetchOrder.func10 (6 handlers)
DELETE /api/order/cancel/:orderId --> src/pkg/module/order.SetupRoute.CancelOrder.func12 (6 handlers)
PUT    /api/order/status/item/:itemId --> src/pkg/module/order.SetupRoute.UpdateItemStatus.func15 (7 handlers)
POST   /api/review/add           --> src/pkg/module/review.SetupRouter.AddReview.func2 (6 handlers)
✅ GET    /api/review               --> src/pkg/module/review.SetupRouter.GetAllReviews.func3 (5 handlers)

✅ GET    /api/review/:slug         --> src/pkg/module/review.SetupRouter.GetProductReviewsBySlug.func4 (5 handlers)
PUT    /api/review/:id           --> src/pkg/module/review.SetupRouter.UpdateReview.func6 (6 handlers)
✅ PUT    /api/review/approve/:reviewId --> src/pkg/module/review.SetupRouter.ApproveReview.func9 (7 handlers)
PUT    /api/review/reject/:reviewId --> src/pkg/module/review.SetupRouter.ApproveReview.func12 (7 handlers)
DELETE /api/review/delete/:id    --> src/pkg/module/review.SetupRouter.DeleteReview.func14 (6 handlers)
POST   /api/payment/webhook      --> src/pkg/module/payment.SetupRouter.handleRazorPayWebhook.func1 (5 handlers)