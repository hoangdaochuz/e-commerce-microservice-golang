"use client";

import React, { createContext, useContext, useState, useEffect } from "react";
import { User, CartItem } from "../types/patient";
import { authService, LoginRequest, LoginResponse } from "@/services/authService";

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  cart: CartItem[];
  cartCount: number;
  addToCart: (item: CartItem) => void;
  removeFromCart: (productId: string) => void;
  clearCart: () => void;
  handleSignIn: (request: LoginRequest) => Promise<LoginResponse>;
  handleSignOut: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [cart, setCart] = useState<CartItem[]>([]);


  useEffect(() => {
    (async () => {
      const claims = await authService.getMe()
      if (claims) {
        setIsAuthenticated(!!claims)
        setUser(claims)
      }
    })()
  }, []);

  const handleSignIn = async (request: LoginRequest): Promise<LoginResponse> => {
    const response = await authService.login(request);
    // if (response)
    return response;
  };

  const handleSignOut = async () => {
    await authService.logout()
  };


  const addToCart = (item: CartItem) => {
    const existingItem = cart.find(i => i.product.id === item.product.id);
    let newCart: CartItem[];

    if (existingItem) {
      newCart = cart.map(i =>
        i.product.id === item.product.id
          ? { ...i, quantity: i.quantity + item.quantity }
          : i
      );
    } else {
      newCart = [...cart, item];
    }

    setCart(newCart);
    localStorage.setItem("cart", JSON.stringify(newCart));
  };

  const removeFromCart = (productId: string) => {
    const newCart = cart.filter(item => item.product.id !== productId);
    setCart(newCart);
    localStorage.setItem("cart", JSON.stringify(newCart));
  };

  const clearCart = () => {
    setCart([]);
    localStorage.removeItem("cart");
  };

  const cartCount = cart.reduce((sum, item) => sum + item.quantity, 0);

  return (
    <AuthContext.Provider value={{
      user,
      isAuthenticated,
      cart,
      cartCount,
      addToCart,
      removeFromCart,
      clearCart,
      handleSignIn,
      handleSignOut
    }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}

