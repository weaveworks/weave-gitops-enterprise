import { TerraformObject } from '../../../api/terraform/types.pb';
import { getLastApplied } from '../TerraformListTable';

describe('getLastApplied', () => {
  const obj: TerraformObject = {
    conditions: [
      {
        type: 'Apply',
        status: 'True',
        timestamp: '2021-08-11T14:23:39Z',
      },
    ],
  };

  it('should return the timestamp', () => {
    expect(getLastApplied(obj)).toBe('2021-08-11T14:23:39.000Z');
  });

  it('should returning "-" if no timestamp', () => {
    const obj2: TerraformObject = {
      conditions: [
        {
          type: 'Apply',
          status: 'True',
        },
      ],
    };
    expect(getLastApplied(obj2)).toBe('-');
  });

  it('should returning "-" if no conditions', () => {
    const obj3: TerraformObject = {};
    expect(getLastApplied(obj3)).toBe('-');
  });

  it('should returning "-" if no conditions with type Apply', () => {
    const obj4: TerraformObject = {
      conditions: [
        {
          type: 'Apply1',
          status: 'True',
          timestamp: '2021-08-11T14:23:39Z',
        },
      ],
    };
    expect(getLastApplied(obj4)).toBe('-');
  });

  it('should returning "-" if the timestamp is not valid', () => {
    const obj5: TerraformObject = {
      conditions: [
        {
          type: 'Apply',
          status: 'True',
          timestamp: 'foo',
        },
      ],
    };

    expect(getLastApplied(obj5)).toBe('-');
  });
});
