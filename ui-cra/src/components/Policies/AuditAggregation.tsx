import { BorderFlexBox } from './PolicyStyles';
import { Flex, Severity, Text } from '@weaveworks/weave-gitops';

const AuditAggregation = () => {
  const FlexItem = (n: number, title: any) => (
    <Flex column>
      <Text bold color="black" size="huge">
        {n}
      </Text>
      <Text color="neutral30" size="large">
        {title}
      </Text>
    </Flex>
  );
  const items: any = [
    { title: 'All Violation', n: 800 },
    { title: <Severity severity="high" />, n: 99 },
    { title: <Severity severity="medium" />, n: 67 },
    { title: <Severity severity="low" />, n: 6535 },
  ];

  return (
    <BorderFlexBox center column gap="32">
      <Text semiBold color="neutral30">
        Violations by Severity
      </Text>
      <Flex wide between>
        {items?.map((item: any) => FlexItem(item.n, item.title))}
      </Flex>
    </BorderFlexBox>
  );
};

export default AuditAggregation;
