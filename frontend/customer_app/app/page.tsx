"use client";

import Link from "next/link";
import { useAuth } from "./context/AuthContext";
import { Product } from "./types/patient";

// Demo products data
const featuredProducts: Product[] = [
  {
    id: "1",
    name: "Wireless Headphones",
    price: 99.99,
    image: "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=500&h=500&fit=crop",
    category: "Electronics",
    rating: 4.5,
    inStock: true,
  },
  {
    id: "2",
    name: "Smart Watch",
    price: 249.99,
    image: "https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=500&h=500&fit=crop",
    category: "Electronics",
    rating: 4.8,
    inStock: true,
  },
  {
    id: "3",
    name: "Running Shoes",
    price: 79.99,
    image: "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=500&h=500&fit=crop",
    category: "Fashion",
    rating: 4.6,
    inStock: true,
  },
  {
    id: "4",
    name: "Laptop Backpack",
    price: 49.99,
    image: "https://images.unsplash.com/photo-1553062407-98eeb64c6a62?w=500&h=500&fit=crop",
    category: "Accessories",
    rating: 4.4,
    inStock: true,
  },
  {
    id: "5",
    name: "Coffee Maker",
    price: 129.99,
    image: "https://images.unsplash.com/photo-1517668808822-9ebb02f2a0e6?w=500&h=500&fit=crop",
    category: "Home",
    rating: 4.7,
    inStock: true,
  },
  {
    id: "6",
    name: "Yoga Mat",
    price: 29.99,
    image: "https://images.unsplash.com/photo-1601925260368-ae2f83cf8b7f?w=500&h=500&fit=crop",
    category: "Sports",
    rating: 4.3,
    inStock: true,
  },
];

export default function Home() {
  const { user, isAuthenticated, addToCart } = useAuth();

  // const handleDemoLogin = () => {
  //   const demoUser = {
  //     id: "1",
  //     name: "John Doe",
  //     email: "john.doe@example.com",
  //     phone: "+1 (555) 123-4567",
  //     address: "123 Main St, City, State 12345",
  //     joinedDate: "2024-01-15",
  //   };
  //   login(demoUser);
  // };

  const handleAddToCart = (product: Product) => {
    addToCart({ product, quantity: 1 });
  };

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-900">
      {/* Hero Section */}
      <section className="bg-gradient-to-r from-blue-600 to-purple-600 text-white py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex flex-col md:flex-row items-center justify-between gap-12">
            <div className="flex-1 text-center md:text-left">
              {isAuthenticated && user ? (
                <>
                  <div className="flex items-center gap-4 mb-4">
                    <div className="w-16 h-16 rounded-full bg-white/20 backdrop-blur flex items-center justify-center text-white font-bold text-2xl">
                      {user.FirstName.charAt(0).toUpperCase()}
                    </div>
                    <div>
                      <p className="text-white/80 text-sm">Welcome back,</p>
                      <h2 className="text-2xl font-bold">{user.FirstName} {user.LastName}!</h2>
                    </div>
                  </div>
                  <h1 className="text-5xl font-bold mb-6">
                    Discover Amazing Products
                  </h1>
                </>
              ) : (
                <>
                  <h1 className="text-5xl font-bold mb-6">
                    Shop the Latest Trends
                  </h1>
                  <p className="text-xl mb-8 opacity-90">
                    Discover amazing products with unbeatable prices and fast shipping
                  </p>
                  {/* <button
                    onClick={handleDemoLogin}
                    className="px-8 py-4 bg-white text-blue-600 rounded-lg hover:bg-zinc-100 transition-colors font-semibold text-lg shadow-lg"
                  >
                    Sign In to Start Shopping
                  </button> */}
                </>
              )}
            </div>
            <div className="flex-1 grid grid-cols-2 gap-4">
              <div className="bg-white/10 backdrop-blur rounded-2xl p-6 text-center">
                <div className="text-4xl font-bold mb-2">10K+</div>
                <div className="text-white/80">Products</div>
              </div>
              <div className="bg-white/10 backdrop-blur rounded-2xl p-6 text-center">
                <div className="text-4xl font-bold mb-2">50K+</div>
                <div className="text-white/80">Customers</div>
              </div>
              <div className="bg-white/10 backdrop-blur rounded-2xl p-6 text-center">
                <div className="text-4xl font-bold mb-2">24/7</div>
                <div className="text-white/80">Support</div>
              </div>
              <div className="bg-white/10 backdrop-blur rounded-2xl p-6 text-center">
                <div className="text-4xl font-bold mb-2">Free</div>
                <div className="text-white/80">Shipping</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Categories */}
      <section className="py-12 bg-white dark:bg-zinc-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <h2 className="text-3xl font-bold text-zinc-900 dark:text-white mb-8">
            Shop by Category
          </h2>
          <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
            {["Electronics", "Fashion", "Home", "Sports", "Books", "Toys"].map((category) => (
              <Link
                key={category}
                href={`/category/${category.toLowerCase()}`}
                className="bg-zinc-100 dark:bg-zinc-700 rounded-xl p-6 text-center hover:shadow-lg transition-shadow group"
              >
                <div className="w-16 h-16 mx-auto mb-4 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center text-white text-2xl group-hover:scale-110 transition-transform">
                  {category.charAt(0)}
                </div>
                <h3 className="font-semibold text-zinc-900 dark:text-white">
                  {category}
                </h3>
              </Link>
            ))}
          </div>
        </div>
      </section>

      {/* Featured Products */}
      <section className="py-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center mb-8">
            <h2 className="text-3xl font-bold text-zinc-900 dark:text-white">
              Featured Products
            </h2>
            <Link
              href="/products"
              className="text-blue-600 dark:text-blue-400 hover:underline font-medium"
            >
              View All →
            </Link>
          </div>
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {featuredProducts.map((product) => (
              <div
                key={product.id}
                className="bg-white dark:bg-zinc-800 rounded-2xl shadow-lg overflow-hidden hover:shadow-xl transition-shadow group"
              >
                <div className="relative overflow-hidden aspect-square">
                  <img
                    src={product.image}
                    alt={product.name}
                    className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-300"
                  />
                  {product.inStock && (
                    <span className="absolute top-4 right-4 bg-green-500 text-white px-3 py-1 rounded-full text-xs font-semibold">
                      In Stock
                    </span>
                  )}
                </div>
                <div className="p-6">
                  <div className="text-sm text-zinc-600 dark:text-zinc-400 mb-2">
                    {product.category}
                  </div>
                  <h3 className="text-xl font-semibold text-zinc-900 dark:text-white mb-2">
                    {product.name}
                  </h3>
                  <div className="flex items-center gap-2 mb-4">
                    <div className="flex items-center">
                      {[...Array(5)].map((_, i) => (
                        <svg
                          key={i}
                          className={`w-4 h-4 ${i < Math.floor(product.rating)
                            ? "text-yellow-400"
                            : "text-zinc-300 dark:text-zinc-600"
                            }`}
                          fill="currentColor"
                          viewBox="0 0 20 20"
                        >
                          <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                        </svg>
                      ))}
                    </div>
                    <span className="text-sm text-zinc-600 dark:text-zinc-400">
                      {product.rating}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-2xl font-bold text-zinc-900 dark:text-white">
                      ${product.price}
                    </span>
                    <button
                      onClick={() => handleAddToCart(product)}
                      className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium"
                    >
                      Add to Cart
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Features Banner */}
      <section className="py-16 bg-white dark:bg-zinc-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid md:grid-cols-4 gap-8">
            <div className="text-center">
              <div className="w-16 h-16 mx-auto mb-4 bg-blue-100 dark:bg-blue-900 rounded-full flex items-center justify-center">
                <svg className="w-8 h-8 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
                </svg>
              </div>
              <h3 className="font-semibold text-zinc-900 dark:text-white mb-2">
                Free Shipping
              </h3>
              <p className="text-sm text-zinc-600 dark:text-zinc-400">
                On orders over $50
              </p>
            </div>
            <div className="text-center">
              <div className="w-16 h-16 mx-auto mb-4 bg-purple-100 dark:bg-purple-900 rounded-full flex items-center justify-center">
                <svg className="w-8 h-8 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                </svg>
              </div>
              <h3 className="font-semibold text-zinc-900 dark:text-white mb-2">
                Secure Payment
              </h3>
              <p className="text-sm text-zinc-600 dark:text-zinc-400">
                100% secure transactions
              </p>
            </div>
            <div className="text-center">
              <div className="w-16 h-16 mx-auto mb-4 bg-green-100 dark:bg-green-900 rounded-full flex items-center justify-center">
                <svg className="w-8 h-8 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
                </svg>
              </div>
              <h3 className="font-semibold text-zinc-900 dark:text-white mb-2">
                Easy Returns
              </h3>
              <p className="text-sm text-zinc-600 dark:text-zinc-400">
                30-day return policy
              </p>
            </div>
            <div className="text-center">
              <div className="w-16 h-16 mx-auto mb-4 bg-orange-100 dark:bg-orange-900 rounded-full flex items-center justify-center">
                <svg className="w-8 h-8 text-orange-600 dark:text-orange-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192l-3.536 3.536M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z" />
                </svg>
              </div>
              <h3 className="font-semibold text-zinc-900 dark:text-white mb-2">
                24/7 Support
              </h3>
              <p className="text-sm text-zinc-600 dark:text-zinc-400">
                Always here to help
              </p>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
