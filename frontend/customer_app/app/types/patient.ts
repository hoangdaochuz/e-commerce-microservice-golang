export interface User {
  Id: string;
  ExternalUserId: string;
  Username: string;
  Email: string;
  FirstName: string;
  LastName: string;
  Gender: string;
}

export interface Product {
  id: string;
  name: string;
  price: number;
  image: string;
  category: string;
  rating: number;
  inStock: boolean;
}

export interface CartItem {
  product: Product;
  quantity: number;
}

export interface Order {
  id: string;
  date: string;
  total: number;
  status: "processing" | "shipped" | "delivered" | "cancelled";
  items: CartItem[];
}

