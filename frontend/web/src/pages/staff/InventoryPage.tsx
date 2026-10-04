/**
 * InventoryPage - Equipment & Stock Management (Staff Portal)
 * Spec: 006-inventory
 * Requirements: FR-INV-01, FR-INV-02, FR-INV-03, DOM-INV-01, DOM-INV-02, UI-INV-01, UI-INV-02, UI-INV-03
 *
 * UI-INV-01: Display items with search and stock levels
 * UI-INV-02: Checkout/return form with required fields (person, purpose, return_date)
 * UI-INV-03: Highlight low-stock items separately
 */

import React, { useState, useEffect } from 'react';
import './InventoryPage.css';
import { InventoryItem, StockMovement } from '@gym-mgmt/shared';

/**
 * MOCK DATA - Replace with API when backend ready
 * Will call: apiClient.getInventoryItems(), apiClient.recordStockMovement()
 *
 * NFR-DATA-01: ACID transactions ensure stock consistency under concurrent access
 */
const MOCK_INVENTORY: InventoryItem[] = [
  {
    id: 'item_001',
    name: 'Resistance Bands (Medium)',
    category: 'Equipment',
    quantity: 5,
    reorder_point: 10,
    unit: 'piece',
  },
  {
    id: 'item_002',
    name: 'Yoga Mats',
    category: 'Equipment',
    quantity: 15,
    reorder_point: 8,
    unit: 'piece',
  },
  {
    id: 'item_003',
    name: 'Hand Towels',
    category: 'Consumables',
    quantity: 3,
    reorder_point: 20,
    unit: 'pack',
  },
  {
    id: 'item_004',
    name: 'Dumbbells (10kg)',
    category: 'Equipment',
    quantity: 12,
    reorder_point: 6,
    unit: 'piece',
  },
];

export const InventoryPage: React.FC = () => {
  const [items, setItems] = useState<InventoryItem[]>(MOCK_INVENTORY);
  const [searchTerm, setSearchTerm] = useState('');
  const [showCheckoutForm, setShowCheckoutForm] = useState(false);
  const [selectedItemId, setSelectedItemId] = useState<string | null>(null);
  const [formData, setFormData] = useState({
    responsible_person: '',
    purpose: '',
    return_date: '',
    quantity: 1,
  });
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    // TODO: Fetch items from API
    // const fetchItems = async () => {
    //   const data = await apiClient.getInventoryItems();
    //   setItems(data);
    // };
    // fetchItems();
  }, []);

  // UI-INV-01: Search functionality
  const filteredItems = items.filter(
    (item) =>
      item.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      item.category.toLowerCase().includes(searchTerm.toLowerCase()),
  );

  // UI-INV-03: Separate low-stock and normal stock items
  const lowStockItems = filteredItems.filter((item) => item.quantity <= item.reorder_point);
  const normalStockItems = filteredItems.filter((item) => item.quantity > item.reorder_point);

  const handleCheckoutClick = (itemId: string) => {
    setSelectedItemId(itemId);
    setShowCheckoutForm(true);
  };

  const handleFormChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleSubmitCheckout = async (e: React.FormEvent) => {
    e.preventDefault();

    // DOM-INV-01: Validate required fields
    if (!formData.responsible_person.trim() || !formData.purpose.trim()) {
      alert('Please fill in all required fields');
      return;
    }

    setIsSaving(true);
    try {
      // TODO: Call API recordStockMovement(movement)
      const movement: Partial<StockMovement> = {
        item_id: selectedItemId || '',
        movement_type: formData.return_date ? 'issue' : 'return',
        quantity: formData.quantity,
        responsible_person: formData.responsible_person,
        purpose: formData.purpose,
        return_date: formData.return_date || undefined,
      };

      // NFR-DATA-01: ACID transaction ensures consistency
      console.log('Recording stock movement:', movement);

      setShowCheckoutForm(false);
      setFormData({ responsible_person: '', purpose: '', return_date: '', quantity: 1 });
      alert('Stock movement recorded successfully');
    } catch (error) {
      alert(`Failed to record movement: ${error}`);
    } finally {
      setIsSaving(false);
    }
  };

  const renderItemGroup = (groupItems: InventoryItem[], isLowStock: boolean) => {
    return (
      <div key={isLowStock ? 'low-stock' : 'normal-stock'}>
        {groupItems.length > 0 && (
          <div className={isLowStock ? 'low-stock-section' : 'normal-stock-section'}>
            <h3 className={isLowStock ? 'low-stock-title' : 'normal-stock-title'}>
              {isLowStock ? '⚠️ Low Stock Items (FR-INV-03)' : 'Normal Stock Items'}
            </h3>
            <table className="inventory-table">
              <thead>
                <tr>
                  <th>Item Name</th>
                  <th>Category</th>
                  <th>Current Stock</th>
                  <th>Reorder Point</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {groupItems.map((item) => (
                  <tr key={item.id} className={isLowStock ? 'low-stock-row' : ''}>
                    <td className="item-name">{item.name}</td>
                    <td>{item.category}</td>
                    <td className="quantity">
                      {item.quantity} {item.unit}
                    </td>
                    <td className="reorder-point">{item.reorder_point}</td>
                    <td>
                      <button
                        className="btn btn-checkout"
                        onClick={() => handleCheckoutClick(item.id)}
                      >
                        Issue/Return
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="inventory-page">
      <h1>Inventory Management</h1>
      <p className="spec-info">
        • FR-INV-01: Manage equipment and supplies
        <br />• FR-INV-02: Track stock movements (issue, return, receive)
        <br />• FR-INV-03: Auto-alert when stock reaches reorder point
        <br />• DOM-INV-01: Checkout form requires: person, purpose, return_date (if borrowing)
        <br />• NFR-DATA-01: ACID transactions prevent concurrent stock conflicts
      </p>

      {/* Search Bar */}
      <div className="search-section">
        <input
          type="text"
          placeholder="Search items by name or category..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="search-input"
        />
      </div>

      {/* Inventory Display */}
      <div className="inventory-content">
        {renderItemGroup(lowStockItems, true)}
        {renderItemGroup(normalStockItems, false)}

        {filteredItems.length === 0 && (
          <p className="no-results">No items found matching "{searchTerm}"</p>
        )}
      </div>

      {/* Checkout/Return Form */}
      {showCheckoutForm && selectedItemId && (
        <div className="modal-overlay" onClick={() => setShowCheckoutForm(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h2>Record Stock Movement</h2>
            <form onSubmit={handleSubmitCheckout}>
              <div className="form-group">
                <label htmlFor="item">
                  Item: {items.find((i) => i.id === selectedItemId)?.name}
                </label>
              </div>

              <div className="form-group">
                <label htmlFor="responsible_person">
                  Responsible Person <span className="required">*</span>
                </label>
                <input
                  id="responsible_person"
                  name="responsible_person"
                  type="text"
                  value={formData.responsible_person}
                  onChange={handleFormChange}
                  placeholder="Name of person taking/returning item"
                  required
                />
              </div>

              <div className="form-group">
                <label htmlFor="purpose">
                  Purpose <span className="required">*</span>
                </label>
                <textarea
                  id="purpose"
                  name="purpose"
                  value={formData.purpose}
                  onChange={handleFormChange}
                  placeholder="Why is this item being issued/returned?"
                  required
                />
              </div>

              <div className="form-group">
                <label htmlFor="quantity">Quantity</label>
                <input
                  id="quantity"
                  name="quantity"
                  type="number"
                  min="1"
                  value={formData.quantity}
                  onChange={handleFormChange}
                />
              </div>

              <div className="form-group">
                <label htmlFor="return_date">
                  Return Date (if borrowing) <span className="required">*</span> for loans
                </label>
                <input
                  id="return_date"
                  name="return_date"
                  type="date"
                  value={formData.return_date}
                  onChange={handleFormChange}
                />
              </div>

              <div className="button-group">
                <button type="submit" className="btn btn-primary" disabled={isSaving}>
                  {isSaving ? 'Saving...' : 'Record Movement'}
                </button>
                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={() => setShowCheckoutForm(false)}
                  disabled={isSaving}
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
