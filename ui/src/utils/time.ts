import moment from 'moment';
import { Duration } from '../types/global';

export const asSeconds = (duration: Duration) => {
  let result = moment.duration(Number(duration));
  const parsed = /([0-9]+)([a-z]+)/.exec(duration);
  if (parsed) {
    const input = Number(parsed[1]); // will be `"5"` if the input is `"5m"`
    const unit = parsed[2] as moment.unitOfTime.DurationConstructor; // will be `"m"` if the input is `"5m"`
    result = moment.duration(input, unit);
  }
  return result.asSeconds();
};
