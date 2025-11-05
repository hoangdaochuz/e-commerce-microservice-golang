"use client";

import { useAuth } from "../context/AuthContext";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { Order } from "../types/patient";

// Demo orders data
const demoOrders: Order[] = [
  {
    id: "ORD-001",
    date: "2024-10-28",
    total: 179.98,
    status: "delivered",
    items: [
      {
        product: {
          id: "1",
          name: "Wireless Headphones",
          price: 99.99,
          image: "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=200&h=200&fit=crop",
          category: "Electronics",
          rating: 4.5,
          inStock: true,
        },
        quantity: 1,
      },
      {
        product: {
          id: "4",
          name: "Laptop Backpack",
          price: 49.99,
          image: "https://images.unsplash.com/photo-1553062407-98eeb64c6a62?w=200&h=200&fit=crop",
          category: "Accessories",
          rating: 4.4,
          inStock: true,
        },
        quantity: 1,
      },
    ],
  },
  {
    id: "ORD-002",
    date: "2024-10-25",
    total: 249.99,
    status: "shipped",
    items: [
      {
        product: {
          id: "2",
          name: "Smart Watch",
          price: 249.99,
          image: "https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=200&h=200&fit=crop",
          category: "Electronics",
          rating: 4.8,
          inStock: true,
        },
        quantity: 1,
      },
    ],
  },
  {
    id: "ORD-003",
    date: "2024-10-20",
    total: 159.98,
    status: "processing",
    items: [
      {
        product: {
          id: "5",
          name: "Coffee Maker",
          price: 129.99,
          image: "https://images.unsplash.com/photo-1517668808822-9ebb02f2a0e6?w=200&h=200&fit=crop",
          category: "Home",
          rating: 4.7,
          inStock: true,
        },
        quantity: 1,
      },
      {
        product: {
          id: "6",
          name: "Yoga Mat",
          price: 29.99,
          image: "https://images.unsplash.com/photo-1601925260368-ae2f83cf8b7f?w=200&h=200&fit=crop",
          category: "Sports",
          rating: 4.3,
          inStock: true,
        },
        quantity: 1,
      },
    ],
  },
];

export default function ProfilePage() {
  const { user, isAuthenticated } = useAuth();
  const router = useRouter();
  const [activeTab, setActiveTab] = useState<"profile" | "orders" | "wishlist">("profile");

  // useEffect(() => {
  //   if (!isAuthenticated) {
  //     router.push("/");
  //   }
  // }, [isAuthenticated, router]);

  if (!isAuthenticated || !user) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-zinc-900">
        <div className="text-center">
          <p className="text-zinc-600 dark:text-zinc-400">Loading...</p>
        </div>
      </div>
    );
  }

  const getStatusColor = (status: Order["status"]) => {
    switch (status) {
      case "delivered":
        return "bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300";
      case "shipped":
        return "bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300";
      case "processing":
        return "bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300";
      case "cancelled":
        return "bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300";
      default:
        return "bg-zinc-100 text-zinc-700 dark:bg-zinc-700 dark:text-zinc-300";
    }
  };

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-900">
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        {/* Profile Header */}
        <div className="bg-gradient-to-r from-blue-600 to-purple-600 rounded-2xl shadow-lg p-8 mb-8 text-white">
          <div className="flex items-center gap-6">
            <div className="w-24 h-24 rounded-full bg-white/20 backdrop-blur flex items-center justify-center text-white font-bold text-4xl shadow-xl flex-shrink-0">
              {user.FirstName.charAt(0).toUpperCase()}
            </div>
            <div className="flex-1">
              <h1 className="text-4xl font-bold mb-2">{user.FirstName} {user.LastName}</h1>
              <p className="text-white/90 mb-1">{user.Email}</p>
            </div>
            <button className="px-6 py-3 bg-white/20 backdrop-blur hover:bg-white/30 rounded-lg transition-colors font-medium">
              Edit Profile
            </button>
          </div>
        </div>

        {/* Tabs */}
        <div className="bg-white dark:bg-zinc-800 rounded-2xl shadow-lg mb-8">
          <div className="flex border-b border-zinc-200 dark:border-zinc-700">
            <button
              onClick={() => setActiveTab("profile")}
              className={`flex-1 py-4 px-6 font-medium transition-colors ${activeTab === "profile"
                ? "text-blue-600 border-b-2 border-blue-600"
                : "text-zinc-600 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-white"
                }`}
            >
              Profile Information
            </button>
            <button
              onClick={() => setActiveTab("orders")}
              className={`flex-1 py-4 px-6 font-medium transition-colors ${activeTab === "orders"
                ? "text-blue-600 border-b-2 border-blue-600"
                : "text-zinc-600 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-white"
                }`}
            >
              Order History
            </button>
            <button
              onClick={() => setActiveTab("wishlist")}
              className={`flex-1 py-4 px-6 font-medium transition-colors ${activeTab === "wishlist"
                ? "text-blue-600 border-b-2 border-blue-600"
                : "text-zinc-600 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-white"
                }`}
            >
              Wishlist
            </button>
          </div>

          <div className="p-8">
            {/* Profile Information Tab */}
            {activeTab === "profile" && (
              <div className="space-y-6">
                <div className="grid md:grid-cols-2 gap-6">
                  <div>
                    <label className="block text-sm font-medium text-zinc-600 dark:text-zinc-400 mb-2">
                      Full Name
                    </label>
                    <input
                      type="text"
                      value={user.FirstName}
                      readOnly
                      className="w-full px-4 py-3 bg-zinc-50 dark:bg-zinc-700 border border-zinc-200 dark:border-zinc-600 rounded-lg text-zinc-900 dark:text-white"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-zinc-600 dark:text-zinc-400 mb-2">
                      Email Address
                    </label>
                    <input
                      type="email"
                      value={user.Email}
                      readOnly
                      className="w-full px-4 py-3 bg-zinc-50 dark:bg-zinc-700 border border-zinc-200 dark:border-zinc-600 rounded-lg text-zinc-900 dark:text-white"
                    />
                  </div>
                  {user.Gender && (
                    <div>
                      <label className="block text-sm font-medium text-zinc-600 dark:text-zinc-400 mb-2">
                        Gender
                      </label>
                      <input
                        type="text"
                        value={user.Gender}
                        readOnly
                        className="w-full px-4 py-3 bg-zinc-50 dark:bg-zinc-700 border border-zinc-200 dark:border-zinc-600 rounded-lg text-zinc-900 dark:text-white"
                      />
                    </div>
                  )}
                  {/* <div className="md:col-span-2">
                    <label className="block text-sm font-medium text-zinc-600 dark:text-zinc-400 mb-2">
                      Shipping Address
                    </label>
                    <input
                      type="text"
                      value={user.Gender}
                      readOnly
                      className="w-full px-4 py-3 bg-zinc-50 dark:bg-zinc-700 border border-zinc-200 dark:border-zinc-600 rounded-lg text-zinc-900 dark:text-white"
                    />
                  </div> */}
                </div>
                <div className="flex gap-4 pt-4">
                  <button className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium">
                    Update Profile
                  </button>
                  <button className="px-6 py-3 bg-zinc-200 dark:bg-zinc-700 text-zinc-900 dark:text-white rounded-lg hover:bg-zinc-300 dark:hover:bg-zinc-600 transition-colors font-medium">
                    Change Password
                  </button>
                </div>
              </div>
            )}

            {/* Order History Tab */}
            {activeTab === "orders" && (
              <div className="space-y-6">
                {demoOrders.map((order) => (
                  <div
                    key={order.id}
                    className="border border-zinc-200 dark:border-zinc-700 rounded-xl p-6 hover:shadow-md transition-shadow"
                  >
                    <div className="flex justify-between items-start mb-4">
                      <div>
                        <h3 className="text-lg font-semibold text-zinc-900 dark:text-white">
                          Order {order.id}
                        </h3>
                        <p className="text-sm text-zinc-600 dark:text-zinc-400">
                          Placed on {new Date(order.date).toLocaleDateString("en-US", { year: "numeric", month: "long", day: "numeric" })}
                        </p>
                      </div>
                      <span
                        className={`px-4 py-2 rounded-full text-sm font-medium capitalize ${getStatusColor(order.status)}`}
                      >
                        {order.status}
                      </span>
                    </div>
                    <div className="space-y-3 mb-4">
                      {order.items.map((item) => (
                        <div key={item.product.id} className="flex gap-4">
                          <img
                            src={item.product.image}
                            alt={item.product.name}
                            className="w-16 h-16 rounded-lg object-cover"
                          />
                          <div className="flex-1">
                            <h4 className="font-medium text-zinc-900 dark:text-white">
                              {item.product.name}
                            </h4>
                            <p className="text-sm text-zinc-600 dark:text-zinc-400">
                              Quantity: {item.quantity}
                            </p>
                          </div>
                          <div className="text-right">
                            <p className="font-semibold text-zinc-900 dark:text-white">
                              ${item.product.price}
                            </p>
                          </div>
                        </div>
                      ))}
                    </div>
                    <div className="flex justify-between items-center pt-4 border-t border-zinc-200 dark:border-zinc-700">
                      <span className="text-lg font-bold text-zinc-900 dark:text-white">
                        Total: ${order.total.toFixed(2)}
                      </span>
                      <div className="flex gap-3">
                        <button className="px-4 py-2 text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-lg transition-colors font-medium">
                          View Details
                        </button>
                        {order.status === "delivered" && (
                          <button className="px-4 py-2 text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-lg transition-colors font-medium">
                            Buy Again
                          </button>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* Wishlist Tab */}
            {activeTab === "wishlist" && (
              <div className="text-center py-12">
                <svg
                  className="w-24 h-24 mx-auto mb-4 text-zinc-300 dark:text-zinc-600"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"
                  />
                </svg>
                <h3 className="text-xl font-semibold text-zinc-900 dark:text-white mb-2">
                  Your wishlist is empty
                </h3>
                <p className="text-zinc-600 dark:text-zinc-400 mb-6">
                  Start adding products you love to your wishlist
                </p>
                <button className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium">
                  Browse Products
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Account Stats */}
        <div className="grid md:grid-cols-4 gap-6">
          <div className="bg-white dark:bg-zinc-800 rounded-xl p-6 shadow-md">
            <div className="flex items-center justify-between mb-2">
              <span className="text-zinc-600 dark:text-zinc-400">Total Orders</span>
              <svg className="w-8 h-8 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
              </svg>
            </div>
            <div className="text-3xl font-bold text-zinc-900 dark:text-white">
              {demoOrders.length}
            </div>
          </div>
          <div className="bg-white dark:bg-zinc-800 rounded-xl p-6 shadow-md">
            <div className="flex items-center justify-between mb-2">
              <span className="text-zinc-600 dark:text-zinc-400">Total Spent</span>
              <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div className="text-3xl font-bold text-zinc-900 dark:text-white">
              ${demoOrders.reduce((sum, order) => sum + order.total, 0).toFixed(2)}
            </div>
          </div>
          <div className="bg-white dark:bg-zinc-800 rounded-xl p-6 shadow-md">
            <div className="flex items-center justify-between mb-2">
              <span className="text-zinc-600 dark:text-zinc-400">Wishlist Items</span>
              <svg className="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
              </svg>
            </div>
            <div className="text-3xl font-bold text-zinc-900 dark:text-white">0</div>
          </div>
          <div className="bg-white dark:bg-zinc-800 rounded-xl p-6 shadow-md">
            <div className="flex items-center justify-between mb-2">
              <span className="text-zinc-600 dark:text-zinc-400">Saved</span>
              <svg className="w-8 h-8 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
              </svg>
            </div>
            <div className="text-3xl font-bold text-zinc-900 dark:text-white">$45</div>
          </div>
        </div>
      </main>
    </div>
  );
}
