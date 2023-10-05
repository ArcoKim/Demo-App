import { render, screen } from '@testing-library/react';
import Health from './Health';
import Version from './Version';

test('health test', () => {
  render(<Health />);
  const linkElement = screen.getByText(/ok/i);
  expect(linkElement).toBeInTheDocument();
});

test('version test', () => {
  render(<Version />);
  const linkElement = screen.getByText(/0.0.1/i);
  expect(linkElement).toBeInTheDocument();
});